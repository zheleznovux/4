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
	log             Logger
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
			return errors.New("regexp found no number")
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
	dwc.log = Logger{
		ParentNodeName: dwc.name,
		IsLogOutput:    nt.Log,
	}
	return nil
}

func (dwc *DWordCommander) Name() string {
	return dwc.name
}

func (dwc *DWordCommander) checkValue(dwt *tag.DWordTag) bool {
	condition := true
	for i := range dwc.dwordConditions {
		switch dwc.dwordConditions[i].operator {
		case constants.MORE:
			condition = (dwt.Value() > dwc.dwordConditions[i].valueCondition) && condition
		case constants.LESS:
			condition = (dwt.Value() < dwc.dwordConditions[i].valueCondition) && condition
		case constants.EQUAL:
			condition = (dwt.Value() == dwc.dwordConditions[i].valueCondition) && condition
		case constants.NOT_EQUAL:
			condition = (dwt.Value() != dwc.dwordConditions[i].valueCondition) && condition
		case constants.MORE_EQUAL:
			condition = (dwt.Value() >= dwc.dwordConditions[i].valueCondition) && condition
		case constants.LESS_EQUAL:
			condition = (dwt.Value() <= dwc.dwordConditions[i].valueCondition) && condition
		case constants.BIT:
			condition = ((dwt.Value() & uint32(math.Pow(2, float64(dwc.dwordConditions[i].valueCondition)))) != 0) && condition
		case constants.NOT_BIT:
			condition = ((dwt.Value() & uint32(math.Pow(2, float64(dwc.dwordConditions[i].valueCondition)))) == 0) && condition
		default:
			condition = false
		}
	}

	switch dwc.logic {
	case constants.AND:
		{
			condition = condition && dwt.State()
		}
	case constants.OR:
		{
			condition = condition || dwt.State()
		}
	}

	return condition
}

func (dwc *DWordCommander) StartChecking(quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(dwc.scanPeriod)
	var condition chan bool = make(chan bool)

	wg.Add(1)
	go dwc.startCommand(condition, quit, wg)
	for {
		select {
		case <-quit:
			{
				return
			}
		case <-ticker.C:
			{
				ticker.Stop()
				wt, err := dwc.th_ptr.GetTagByName(dwc.Name())
				if err != nil {
					dwc.log.Write(INFO, err.Error())
					return
				}
				condition <- dwc.checkValue(wt.(*tag.DWordTag))

				ticker.Reset(dwc.scanPeriod)
			}
		}
	}
}

func (wc *DWordCommander) startCommand(condition chan bool, quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	timeBetweenTick := wc.actionTimeout / 5
	tickerToCommand := time.NewTicker(wc.actionTimeout)
	tickerToCommand.Stop()
	tickCount := 0

	var lastCondition bool
	for {
		select {
		case <-quit:
			{
				return
			}
		case <-tickerToCommand.C:
			{
				tickerToCommand.Stop()
				tickCount++
				if tickCount != 5 {
					timeToCommand := wc.actionTimeout - time.Duration(tickCount)*timeBetweenTick
					wc.log.Write(INFO, "Команда "+wc.action+", до завершения таймера: "+timeToCommand.String()+".")

				} else {
					tickCount = 0
					wc.log.Write(INFO, "Запущена команда!")
					err := command(wc.action)
					if err != nil {
						wc.log.Write(ERROR, err.Error())
					}
				}
				tickerToCommand.Reset(timeBetweenTick)
			}
		case v := <-condition:
			{
				if lastCondition != v {
					lastCondition = v
					tickCount = 0
					if v {
						wc.log.Write(INFO, "Запущен таймер команды "+wc.action+", до завершения: "+wc.actionTimeout.String()+".")
						tickerToCommand.Reset(timeBetweenTick)
					} else {
						wc.log.Write(INFO, "Таймер команды остановлен!")
						tickerToCommand.Stop()
					}
				}
			}
		}
	}
}
