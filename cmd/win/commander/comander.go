package commander

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	"zheleznovux.com/modbus-console/cmd/constants"
	tags "zheleznovux.com/modbus-console/cmd/storage"
)

type CommanderInterface interface {
	setup(configuration.NodeTag, *tags.TagsHandler) error
	StartChecking(chan int)
	checkLogic(bool, bool) bool
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
	return "", fmt.Errorf("did not have action")
}

func makeSecond(t float64) time.Duration {
	return time.Duration(t * float64(time.Second))
}

func Setup(nt configuration.NodeTag, th *tags.TagsHandler) (CommanderInterface, error) {
	fmt.Println(nt.DataType)
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
		return nil, fmt.Errorf("did not have data type")
	}
}

func command(c string) error {
	var flag string
	exe := strings.Split(c, " ")

	for i := range exe {
		fmt.Println(exe[i])
	}

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
