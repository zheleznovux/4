package commander

import (
	"errors"
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
	log            Logger
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
	wc.log = Logger{
		ParentNodeName: wc.name,
		IsLogOutput:    nt.Log,
	}
	return nil
}

func (wc *CoilCommander) Name() string {
	return wc.name
}

func (cc *CoilCommander) StartChecking(quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(cc.scanPeriod)
	var condition chan bool = make(chan bool)

	wg.Add(1)
	go cc.startCommand(condition, quit, wg)
	for {
		select {
		case <-quit:
			{
				return
			}
		case <-ticker.C:
			{
				ticker.Stop()
				wt, err := cc.th_ptr.GetTagByName(cc.Name())
				if err != nil {
					cc.log.Write(ERROR, err.Error())
					return
				}
				condition <- cc.checkValue(wt.(*tag.CoilTag))

				ticker.Reset(cc.scanPeriod)
			}
		}
	}
}

func (cc *CoilCommander) checkValue(ct *tag.CoilTag) bool {
	switch cc.logic {
	case constants.AND:
		{
			return (ct.Value() == 1) == cc.valueCondition && ct.State()
		}
	case constants.OR:
		{
			return (ct.Value() == 1) == cc.valueCondition || ct.State()
		}
	default:
		return false
	}
}

func (cc *CoilCommander) startCommand(condition chan bool, quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	timeBetweenTick := cc.actionTimeout / 5
	tickerToCommand := time.NewTicker(cc.actionTimeout)
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
					timeToCommand := cc.actionTimeout - time.Duration(tickCount)*timeBetweenTick
					cc.log.Write(INFO, "Команда "+cc.action+", до завершения таймера: "+timeToCommand.String()+".")
				} else {
					tickCount = 0
					cc.log.Write(INFO, "Запущена команда!")
					err := command(cc.action)
					if err != nil {
						cc.log.Write(ERROR, err.Error())
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
						cc.log.Write(INFO, "Запущен таймер команды "+cc.action+", до исполнения: "+cc.actionTimeout.String()+".")
						tickerToCommand.Reset(timeBetweenTick)
					} else {
						cc.log.Write(INFO, "Таймер команды остановлен!")
						tickerToCommand.Stop()
					}
				}
			}
		}
	}
}
