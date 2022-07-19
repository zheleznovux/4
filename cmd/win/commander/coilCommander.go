package commander

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	storage "zheleznovux.com/modbus-console/cmd/serverStorage"
	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"
	"zheleznovux.com/modbus-console/cmd/serverStorage/tag"
)

type CoilCommander struct {
	name           string
	stateCondition bool
	valueCondition bool
	logic          string
	action         string
	actionTimeout  time.Duration
	scanPeriod     time.Duration
	th_ptr         *storage.Server
}

func (wc *CoilCommander) setup(nt configuration.NodeTag, th *storage.Server) error {
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

func (cc *CoilCommander) StartChecking(quit chan int, wg *sync.WaitGroup) {
	fmt.Println("запущен cc")
	defer wg.Done()
	defer fmt.Println("прекращен cc")

	for {
		select {
		case <-quit:
			{
				return
			}
		default:
			{
				ct, err := cc.th_ptr.GetTagByName(cc.Name())
				if err != nil {
					fmt.Println(cc.Name() + " " + err.Error())
					return
				}
				op1 := cc.checkValue(ct.(*tag.CoilTag))
				if cc.checkLogic(op1, ct.(*tag.CoilTag).State()) {
					cc.startCommand()
				}
				time.Sleep(cc.scanPeriod)
			}
		}
	}
}

func (cc *CoilCommander) checkValue(ct *tag.CoilTag) bool {
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
