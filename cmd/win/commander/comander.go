package commander

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	tags "zheleznovux.com/modbus-console/cmd/storage"
)

//opertor
const (
	EQUAL      = "="
	NOT_EQUAL  = "!="
	MORE_EQUAL = ">="
	LESS_EQUAL = "<="
	MORE       = ">"
	LESS       = "<"
)

//logic
const (
	AND = "and"
	OR  = "or"
)

//action
const (
	SHUTDOWN = "shutdown"
	// REBOOT      = "reboot"
	RESTART     = "restart"
	RUN_PROGRAM = "run"
)

type Commander interface {
	// Setup(configuration.NodeTag) error
	StartChecking(chan int)
	checkLogic(bool, bool) bool
	Name() string
}

type WordCommander struct {
	name           string
	stateCondition bool
	valueCondition uint16
	operator       string
	logic          string
	action         string
	actionTimeout  time.Duration
	scanPeriod     time.Duration
	th_ptr         *tags.TagsHandler
}

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

func (wc *WordCommander) Name() string {
	return wc.name
}
func (wc *CoilCommander) Name() string {
	return wc.name
}

func makeWordValueCondition(s string) (string, uint16, error) {
	re := regexp.MustCompile("[0-9]+")
	matcher := re.FindAllString(s, -1)
	if len(matcher) != 1 {
		return "", 0, fmt.Errorf("did not have match")
	}
	var tmpNumber uint16
	_, err := fmt.Sscanf(matcher[0], "%d", &tmpNumber)
	if err != nil {
		fmt.Println(err)
		return "", 0, fmt.Errorf("did not have number")
	}

	tmpString := strconv.Itoa(int(tmpNumber))
	split := strings.Split(s, tmpString)
	if len(split) != 2 {
		return "", 0, fmt.Errorf("did not have opertor")
	}

	var rtn string
	switch split[0] {
	case EQUAL:
		{
			rtn = EQUAL
		}
	case NOT_EQUAL:
		{
			rtn = NOT_EQUAL
		}
	case MORE_EQUAL:
		{
			rtn = MORE_EQUAL
		}
	case LESS_EQUAL:
		{
			rtn = LESS_EQUAL
		}
	case MORE:
		{
			rtn = MORE
		}
	case LESS:
		{
			rtn = LESS
		}
	default:
		return "", 0, fmt.Errorf("did not have opertor")
	}

	return rtn, tmpNumber, nil
}

func makeAction(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if strings.Contains(trimmed, SHUTDOWN) {
		return SHUTDOWN, nil
	}
	// if strings.Contains(trimmed, REBOOT) {
	// 	return REBOOT, nil
	// }
	if strings.Contains(trimmed, RESTART) {
		return RESTART, nil
	}
	if strings.Contains(trimmed, RUN_PROGRAM+" ") {
		return trimmed, nil
	}
	return "", fmt.Errorf("did not have action")
}

func makeSecond(t float64) time.Duration {
	return time.Duration(t * float64(time.Second))
}

func Setup(nt configuration.NodeTag, th *tags.TagsHandler) (Commander, error) {
	switch nt.DataType {
	case "word":
		{
			var rtn WordCommander
			var err error

			rtn.name = nt.Name
			if nt.Name == "" {
				return nil, fmt.Errorf("did not have name")
			}

			if nt.StateCondition == "good" {
				rtn.stateCondition = true
			} else if nt.StateCondition == "bad" {
				rtn.stateCondition = false
			} else {
				return nil, fmt.Errorf("did not have stateCondition")
			}

			rtn.operator, rtn.valueCondition, err = makeWordValueCondition(nt.ValueCondition)
			if err != nil {
				return nil, err
			}

			if (strings.ToLower(nt.Logic) != AND) && (strings.ToLower(nt.Logic) != OR) {
				return nil, fmt.Errorf("did not have logic")
			} else {
				rtn.logic = nt.Logic
			}

			rtn.action, err = makeAction(nt.Action)
			if err != nil {
				return nil, err
			}

			rtn.actionTimeout = makeSecond(nt.ActionTimeout)
			rtn.scanPeriod = makeSecond(nt.ScanPeriod)

			rtn.th_ptr = th

			return &rtn, nil
		}
	case "coil":
		{
			var rtn CoilCommander
			var err error

			rtn.name = nt.Name
			if nt.Name == "" {
				return nil, fmt.Errorf("did not have name")
			}

			if nt.StateCondition == "good" {
				rtn.stateCondition = true
			} else if nt.StateCondition == "bad" {
				rtn.stateCondition = false
			} else {
				return nil, fmt.Errorf("did not have stateCondition")
			}

			if nt.ValueCondition == "true" {
				rtn.valueCondition = true
			} else if nt.ValueCondition == "false" {
				rtn.valueCondition = false
			} else {
				return nil, fmt.Errorf("did not have valueCondition")
			}

			if (strings.ToLower(nt.Logic) != AND) && (strings.ToLower(nt.Logic) != OR) {
				return nil, fmt.Errorf("did not have logic")
			} else {
				rtn.logic = nt.Logic
			}

			rtn.action, err = makeAction(nt.Action)
			if err != nil {
				return nil, err
			}

			rtn.actionTimeout = makeSecond(nt.ActionTimeout)
			rtn.scanPeriod = makeSecond(nt.ScanPeriod)

			rtn.th_ptr = th
			return &rtn, nil
		}
	default:
		return nil, fmt.Errorf("did not have data type")
	}
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
				op1 := wc.checkValue(wt.(*tags.WordTag))
				if wc.checkLogic(op1, wt.(*tags.WordTag).State) {
					wc.startCommand()
				}
			}
		case <-quit:
			return
		}
		time.Sleep(wc.scanPeriod)
	}
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
				if cc.checkLogic(op1, ct.(*tags.CoilTag).State) {
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
	return (ct.Value == 1) == cc.valueCondition
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

func command(c string) error {
	var flag string
	exe := strings.Split(c, " ")

	for i := range exe {
		fmt.Println(exe[i])
	}

	switch exe[0] {
	case SHUTDOWN:
		{
			flag = "/s"
		}
	case RESTART:
		{
			flag = "/r"
		}
	case RUN_PROGRAM:
		{
			if len(exe) != 2 {
				return fmt.Errorf("len(2) != 2")
			}
			cmd := exec.Command("./" + exe[1])
			err := cmd.Run()
			if err != nil {
				return err
			}
			return nil
		}
	default:
		return fmt.Errorf("invalid command")
	}

	if err := exec.Command("cmd", "/C", "shutdown "+flag+" /t 1").Run(); err != nil {
		return err
	}
	return nil
}

func timer(t time.Duration, start chan bool) {
	time.Sleep(t)
	start <- true
}

func (wc *WordCommander) checkValue(ct *tags.WordTag) bool {
	switch wc.operator {
	case MORE:
		{
			if ct.Value > wc.valueCondition {
				return true
			}
		}
	case LESS:
		{
			if ct.Value < wc.valueCondition {
				return true
			}
		}
	case EQUAL:
		{
			if ct.Value == wc.valueCondition {
				return true
			}
		}
	case NOT_EQUAL:
		{
			if ct.Value != wc.valueCondition {
				return true
			}
		}
	case MORE_EQUAL:
		{
			if ct.Value >= wc.valueCondition {
				return true
			}
		}
	case LESS_EQUAL:
		{
			if ct.Value <= wc.valueCondition {
				return true
			}
		}
	}
	return false //default
}

func (wc *WordCommander) checkLogic(op1 bool, op2 bool) bool { // kak sdelat' pravil'no?
	switch wc.logic {
	case AND:
		{
			return op1 && op2
		}
	case OR:
		{
			return op1 || op2
		}
	default:
		return false
	}
}

func (wc *CoilCommander) checkLogic(op1 bool, op2 bool) bool {
	switch wc.logic {
	case AND:
		{
			return op1 && op2
		}
	case OR:
		{
			return op1 || op2
		}
	default:
		return false
	}
}
