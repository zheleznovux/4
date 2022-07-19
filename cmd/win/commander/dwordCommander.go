package commander

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"

	storage "zheleznovux.com/modbus-console/cmd/serverStorage"
	"zheleznovux.com/modbus-console/cmd/serverStorage/tag"
)

type DWordCommander struct {
	name            string
	stateCondition  bool
	dwordConditions []dwordCondition
	logic           string
	action          string
	actionTimeout   time.Duration
	scanPeriod      time.Duration
	th_ptr          *storage.Server
}

type dwordCondition struct {
	operator       string
	valueCondition uint32
}

func (dwc *DWordCommander) makeWordValueCondition(s string) error {
	re := regexp.MustCompile(`(?P<OPERATOR>>|>=|<|<=|!=|==|bit|!bit)\(?(?P<VALUE>[0-9]+)\)?`)
	matcher := re.FindAllStringSubmatch(s, -1)

	if len(matcher) == 0 {
		return errors.New("regexp found no match")
	}

	for _, m := range matcher {
		var tmpNumber uint32
		_, err := fmt.Sscanf(m[2], "%d", &tmpNumber)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("regexp found no number")
		}

		wordCondition := dwordCondition{
			operator:       m[1],
			valueCondition: tmpNumber}

		dwc.dwordConditions = append(dwc.dwordConditions, wordCondition)
	}
	return nil
}

func (dwc *DWordCommander) setup(nt configuration.NodeTag, th *storage.Server) error {
	dwc.name = nt.Name
	var err error
	if nt.Name == "" {
		return errors.New("did not have name")
	}

	if nt.StateCondition == "good" {
		dwc.stateCondition = true
	} else if nt.StateCondition == "bad" {
		dwc.stateCondition = false
	} else {
		return errors.New("did not have stateCondition")
	}

	err = dwc.makeWordValueCondition(nt.ValueCondition)
	if err != nil {
		return err
	}

	if (strings.ToLower(nt.Logic) != constants.AND) && (strings.ToLower(nt.Logic) != constants.OR) {
		return errors.New("did not have logic")
	} else {
		dwc.logic = nt.Logic
	}

	dwc.action, err = makeAction(nt.Action)
	if err != nil {
		return err
	}

	dwc.actionTimeout = makeSecond(nt.ActionTimeout)
	dwc.scanPeriod = makeSecond(nt.ScanPeriod)

	dwc.th_ptr = th
	return nil
}

func (dwc *DWordCommander) Name() string {
	return dwc.name
}

func (wc *DWordCommander) startCommand() {
	start := make(chan bool)
	go timer(wc.actionTimeout, start)
	fmt.Println("Запущен таймер команды " + wc.action)
	if <-start {
		err := command(wc.action)
		if err != nil {
			fmt.Println(err)
		} else {
			return
		}
	}
}

func (dwc *DWordCommander) checkValue(ct *tag.DWordTag) bool {
	condition := true
	for i := range dwc.dwordConditions {
		switch dwc.dwordConditions[i].operator {
		case constants.MORE:
			condition = (ct.Value() > dwc.dwordConditions[i].valueCondition) && condition
		case constants.LESS:
			condition = (ct.Value() < dwc.dwordConditions[i].valueCondition) && condition
		case constants.EQUAL:
			condition = (ct.Value() == dwc.dwordConditions[i].valueCondition) && condition
		case constants.NOT_EQUAL:
			condition = (ct.Value() != dwc.dwordConditions[i].valueCondition) && condition
		case constants.MORE_EQUAL:
			condition = (ct.Value() >= dwc.dwordConditions[i].valueCondition) && condition
		case constants.LESS_EQUAL:
			condition = (ct.Value() <= dwc.dwordConditions[i].valueCondition) && condition
		case constants.BIT:
			condition = ((ct.Value() & uint32(math.Pow(2, float64(dwc.dwordConditions[i].valueCondition)))) != 0) && condition
		case constants.NOT_BIT:
			condition = ((ct.Value() & uint32(math.Pow(2, float64(dwc.dwordConditions[i].valueCondition)))) == 0) && condition
		default:
			return false
		}
	}
	return condition
}

func (dwc *DWordCommander) StartChecking(quit chan int, wg *sync.WaitGroup) {
	fmt.Println("запущен dwc")
	defer wg.Done()
	defer fmt.Println("прекращен dwc")

	for {
		select {
		case <-quit:
			{
				return
			}
		default:
			{
				wt, err := dwc.th_ptr.GetTagByName(dwc.Name())
				if err != nil {
					fmt.Println(dwc.Name() + " " + err.Error())
					return
				}
				op1 := dwc.checkValue(wt.(*tag.DWordTag))
				if dwc.checkLogic(op1, wt.(*tag.DWordTag).State()) {
					dwc.startCommand()
				}
				time.Sleep(dwc.scanPeriod)

			}
		}
	}
}

func (dwc *DWordCommander) checkLogic(op1 bool, op2 bool) bool {
	switch dwc.logic {
	case constants.AND:
		{
			return op1 && op2
		}
	case constants.OR:
		{
			return op1 || op2
		}
	default:
		return false
	}
}
