package commander

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	"zheleznovux.com/modbus-console/cmd/constants"
	tags "zheleznovux.com/modbus-console/cmd/storage"
)

type CoilCommander struct {
	name           string
	stateCondition bool
	valueCondition bool
	logic          string
	action         string
	actionTimeout  time.Duration
	scanPeriod     time.Duration
	th_ptr         *tags.TagsHandler
}

func (wc *CoilCommander) setup(nt configuration.NodeTag, th *tags.TagsHandler) error {
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

	if nt.ValueCondition == "true" {
		wc.valueCondition = true
	} else if nt.ValueCondition == "false" {
		wc.valueCondition = false
	} else {
		return errors.New("did not have valueCondition")
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

func (wc *CoilCommander) Name() string {
	return wc.name
}

func (cc *CoilCommander) StartChecking(quit chan int) {
	for {
		select {
		default:
			{
				ct, err := cc.th_ptr.GetTagByName(cc.Name())
				if err != nil {
					return
				}
				op1 := cc.checkValue(ct.(*tags.CoilTag))
				if cc.checkLogic(op1, ct.(*tags.CoilTag).State()) {
					cc.startCommand()
				}
			}
		case <-quit:
			return
		}
		time.Sleep(cc.scanPeriod)
	}
}

func (cc *CoilCommander) checkValue(ct *tags.CoilTag) bool {
	return (ct.Value() == 1) == cc.valueCondition
}

func (cc *CoilCommander) startCommand() {
	start := make(chan bool)
	go timer(cc.actionTimeout, start)
	fmt.Println("Запущен таймер команды " + cc.action)
	if <-start {
		err := command(cc.action)
		if err != nil {
			fmt.Println(err)
		} else {
			return
		}
	}
}

func (wc *CoilCommander) checkLogic(op1 bool, op2 bool) bool {
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
