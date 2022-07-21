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

type WordCommander struct {
	name           string
	stateCondition bool
	wordConditions []wordCondition
	logic          string
	action         string
	actionTimeout  time.Duration
	scanPeriod     time.Duration
	log            Logger
	th_ptr         *storage.Server
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

func (wc *WordCommander) setup(nt configuration.NodeTag, th *storage.Server) error {
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
	wc.log = Logger{
		ParentNodeName: wc.name,
		IsLogOutput:    nt.Log,
	}
	return nil
}

func (wc *WordCommander) Name() string {
	return wc.name
}

func (wc *WordCommander) startCommand(condition chan bool, quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(wc.actionTimeout)
	ticker.Stop()

	var lastCondition bool
	for {
		select {
		case <-quit:
			{
				return
			}
		case <-ticker.C:
			{
				ticker.Stop()
				wc.log.Write(INFO, "Запущена команда!")
				err := command(wc.action)
				if err != nil {
					fmt.Println(err)
				}
				ticker.Reset(wc.actionTimeout)
			}
		case v := <-condition:
			{
				fmt.Println(v)
				if lastCondition != v {
					lastCondition = v
					if v {
						ticker.Reset(wc.actionTimeout)
						wc.log.Write(INFO, "Запущен таймер команды, до завершения: "+wc.actionTimeout.String()+".")

					} else {
						ticker.Stop()
						wc.log.Write(INFO, "Таймер команды остановлен!")
					}
				}
			}
		}
	}
}

func (wc *WordCommander) checkValue(ct *tag.WordTag) bool {
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

	switch wc.logic {
	case constants.AND:
		{
			condition = condition && ct.State()
		}
	case constants.OR:
		{
			condition = condition || ct.State()
		}
	}

	return condition
}

func (wc *WordCommander) StartChecking(quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(wc.scanPeriod)
	var condition chan bool = make(chan bool)

	wg.Add(1)
	go wc.startCommand(condition, quit, wg)
	for {
		select {
		case <-quit:
			{
				return
			}
		case <-ticker.C:
			{
				ticker.Stop()
				wt, err := wc.th_ptr.GetTagByName(wc.Name())
				if err != nil {
					wc.log.Write(ERROR, err.Error())
					return
				}
				condition <- wc.checkValue(wt.(*tag.WordTag))

				ticker.Reset(wc.scanPeriod)
			}
		}
	}
}
