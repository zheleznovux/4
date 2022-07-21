package commander

import (
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	storage "zheleznovux.com/modbus-console/cmd/serverStorage"
	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"

	tags "zheleznovux.com/modbus-console/cmd/serverStorage"
)

type Commander struct {
	name           string
	stateCondition bool
	valueCondition []Condition
	logic          string
	action         string
	actionTimeout  time.Duration
	scanPeriod     time.Duration
	log            Logger
	th_ptr         *storage.Server
}

type CommanderInterface interface {
	setup(configuration.NodeTag, *tags.Server) error
	StartChecking(chan struct{}, *sync.WaitGroup)
	Name() string
}

func makeAction(s string) (string, error) {
	trimmed := strings.TrimSpace(s)

	if strings.Contains(trimmed, constants.SHUTDOWN) {
		return constants.SHUTDOWN, nil
	}
	if strings.Contains(trimmed, constants.RESTART) {
		return constants.RESTART, nil
	}
	if strings.Contains(trimmed, constants.RUN_PROGRAM+" ") {
		return trimmed, nil
	}
	return "", errors.New("did not have action")
}

func makeSecond(t float64) time.Duration {
	return time.Duration(t * float64(time.Second))
}

func Setup(nt configuration.NodeTag, th *tags.Server) (CommanderInterface, error) {
	switch nt.DataType {
	case constants.WORD_TYPE:
		{
			var rtn WordCommander
			err := rtn.setup(nt, th)

			return &rtn, err
		}
	case constants.COIL_TYPE:
		{
			var rtn CoilCommander
			err := rtn.setup(nt, th)

			return &rtn, err
		}
	case constants.DWORD_TYPE:
		{
			var rtn DWordCommander
			err := rtn.setup(nt, th)

			return &rtn, err
		}
	default:
		return nil, errors.New("did not have data type")
	}
}

func command(c string) error {
	var flag string
	exe := strings.Split(c, " ")

	switch exe[0] {
	case constants.SHUTDOWN:
		{
			flag = "/s"
		}
	case constants.RESTART:
		{
			flag = "/r"
		}
	case constants.RUN_PROGRAM:
		{
			if len(exe) != 2 {
				return errors.New("len(exe) != 2")
			}
			cmd := exec.Command("./" + exe[1])
			err := cmd.Run()
			if err != nil {
				return err
			}
			return nil
		}
	default:
		return errors.New("invalid command")
	}

	if err := exec.Command("cmd", "/C", "shutdown "+flag+" /t 1").Run(); err != nil {
		return err
	}
	return nil
}
