package commander

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	"zheleznovux.com/modbus-console/cmd/constants"
	tags "zheleznovux.com/modbus-console/cmd/storage"
)

type WordCommander struct {
	name           string
	stateCondition bool
	wordConditions []wordCondition
	logic          string
	action         string
	actionTimeout  time.Duration
	scanPeriod     time.Duration
	th_ptr         *tags.TagsHandler
}

type wordCondition struct {
	operator       string
	valueCondition uint16
}

func (wc *WordCommander) makeWordValueCondition(s string) error {
	re := regexp.MustCompile(`(?P<OPERATOR>>|>=|<|<=|!=|==|bit|!bit)\(?(?P<VALUE>[0-9]+)\)?`)
	matcher := re.FindAllStringSubmatch(s, -1)

	if len(matcher) == 0 {
		return errors.New("regexp found no match")
	}

	for _, m := range matcher {
		var tmpNumber uint16
		_, err := fmt.Sscanf(m[2], "%d", &tmpNumber)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("regexp found no number")
		}

		wordCondition := wordCondition{
			operator:       m[1],
			valueCondition: tmpNumber}

		wc.wordConditions = append(wc.wordConditions, wordCondition)
	}
	return nil
}

func (wc *WordCommander) setup(nt configuration.NodeTag, th *tags.TagsHandler) error {
	wc.name = nt.Name
	var err error
	if nt.Name == "" {
		return errors.New("did not have name")
	}

	if nt.StateCondition == "good" {
		wc.stateCondition = true
	} else if nt.StateCondition == "bad" {
		wc.stateCondition = false
	} else {
		return errors.New("did not have stateCondition")
	}

	err = wc.makeWordValueCondition(nt.ValueCondition)
	if err != nil {
		return err
	}

	if (strings.ToLower(nt.Logic) != constants.AND) && (strings.ToLower(nt.Logic) != constants.OR) {
		return errors.New("did not have logic")
	} else {
		wc.logic = nt.Logic
	}

	wc.action, err = makeAction(nt.Action)
	if err != nil {
		return err
	}

	wc.actionTimeout = makeSecond(nt.ActionTimeout)
	wc.scanPeriod = makeSecond(nt.ScanPeriod)

	wc.th_ptr = th
	return nil
}

func (wc *WordCommander) Name() string {
	return wc.name
}

func (wc *WordCommander) startCommand() {
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

func (wc *WordCommander) checkValue(ct *tags.WordTag) bool {
	condition := true
	for i := range wc.wordConditions {
		switch wc.wordConditions[i].operator {
		case constants.MORE:
			condition = (ct.Value() > wc.wordConditions[i].valueCondition) && condition
		case constants.LESS:
			condition = (ct.Value() < wc.wordConditions[i].valueCondition) && condition
		case constants.EQUAL:
			condition = (ct.Value() == wc.wordConditions[i].valueCondition) && condition
		case constants.NOT_EQUAL:
			condition = (ct.Value() != wc.wordConditions[i].valueCondition) && condition
		case constants.MORE_EQUAL:
			condition = (ct.Value() >= wc.wordConditions[i].valueCondition) && condition
		case constants.LESS_EQUAL:
			condition = (ct.Value() <= wc.wordConditions[i].valueCondition) && condition
		case constants.BIT:
			condition = ((ct.Value() & uint16(math.Pow(2, float64(wc.wordConditions[i].valueCondition)))) != 0) && condition
		case constants.NOT_BIT:
			condition = ((ct.Value() & uint16(math.Pow(2, float64(wc.wordConditions[i].valueCondition)))) == 0) && condition
		default:
			return false
		}
	}
	return condition
}

func (wc *WordCommander) StartChecking(quit chan int) {
	for {
		select {
		default:
			{
				wt, err := wc.th_ptr.GetTagByName(wc.Name())
				if err != nil {
					return
				}
				fmt.Println(wt)
				op1 := wc.checkValue(wt.(*tags.WordTag))
				if wc.checkLogic(op1, wt.(*tags.WordTag).State()) {
					wc.startCommand()
				}
			}
		case <-quit:
			return
		}
		time.Sleep(wc.scanPeriod)
	}
}

func (wc WordCommander) checkLogic(op1 bool, op2 bool) bool {
	switch wc.logic {
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
