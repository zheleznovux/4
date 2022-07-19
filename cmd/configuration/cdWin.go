package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"
)

type ConfigurationDataWin struct {
	NODES []NodeTag
}

type NodeTag struct {
	Name           string
	DataType       string
	StateCondition string
	ValueCondition string
	Logic          string
	Action         string
	ActionTimeout  float64
	ScanPeriod     float64
}

func (tn *ConfigurationDataWin) Setup(c *ConfigHandler) error {
	content, err := os.ReadFile(c.fileName)
	if err != nil {
		return err
	}

	var tmpTN ConfigurationDataWin
	err = json.Unmarshal(content, &tmpTN)
	if err != nil {
		return err
	}

	// аналогично с cdApp
	for i := 0; i < len(tmpTN.NODES); i++ {
		k := 0
		j := i + 1
		for ; j < len(tmpTN.NODES); j++ {
			if tmpTN.NODES[i].Name != tmpTN.NODES[j].Name {
				k++
			}
		}

		if (j - i - 1) == k {
			var tmpNodeTag NodeTag
			tmpNodeTag.Name, err = verifyName(tmpTN.NODES[i].Name) // проверка из cdApp
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNodeTag.DataType, err = verifyTagDataType(tmpTN.NODES[i].DataType) // проверка из cdApp
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNodeTag.StateCondition, err = verifyStateCondition(tmpTN.NODES[i].StateCondition)
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNodeTag.ValueCondition, err = verifyValueCondition(tmpTN.NODES[i].ValueCondition)
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNodeTag.Logic, err = verifyLogic(tmpTN.NODES[i].Logic)
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNodeTag.Action, err = verifyAction(tmpTN.NODES[i].Action)
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNodeTag.ActionTimeout, err = verifyActionTimeout(tmpTN.NODES[i].ActionTimeout)
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNodeTag.ScanPeriod, err = verifyTagScanPeriod(tmpTN.NODES[i].ScanPeriod) // проверка из cdApp
			if err != nil {
				fmt.Println("Win config: Tag skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}

			// если прошел проверку добавляем
			tn.NODES = append(tn.NODES, tmpNodeTag)
		}
	}

	return nil
}

func verifyStateCondition(state string) (string, error) {
	str := strings.TrimSpace(strings.ToLower(state))

	switch str {
	case constants.BAD:
		return str, nil
	case constants.GOOD:
		return str, nil
	default:
		return str, fmt.Errorf("config did not have state")
	}
}

func verifyValueCondition(state string) (string, error) {
	str := strings.TrimSpace(strings.ToLower(state))

	switch str {
	case "true":
		return str, nil
	case "false":
		return str, nil
	case "1":
		return "true", nil
	case "0":
		return "false", nil
	default:
		{
			if len(str) != 0 {
				return str, nil
			} else {
				return str, fmt.Errorf("config did not have state")

			}
		}
	}
}

func verifyLogic(logic string) (string, error) {
	str := strings.TrimSpace(strings.ToLower(logic))

	switch logic {
	case constants.AND:
		return str, nil
	case constants.OR:
		return str, nil
	case "&&":
		return constants.AND, nil
	case "||":
		return constants.OR, nil
	default:
		return str, fmt.Errorf("config did not have logic")
	}
}

func verifyAction(logic string) (string, error) {
	str := strings.TrimSpace(strings.ToLower(logic))

	switch str {
	case constants.SHUTDOWN:
		return constants.SHUTDOWN, nil
	case constants.RESTART:
		return constants.RESTART, nil
	case "/r":
		return constants.RESTART, nil
	case "r":
		return constants.RESTART, nil
	case "-r":
		return constants.RESTART, nil
	case "/s":
		return constants.RESTART, nil
	case "s":
		return constants.RESTART, nil
	case "-s":
		return constants.RESTART, nil
	default:
		{
			if len(str) != 0 {
				return str, nil
			} else {
				return str, fmt.Errorf("config did not have action")

			}
		}
	}
}

func verifyActionTimeout(t float64) (float64, error) {
	if t < 0 {
		return t, fmt.Errorf("config action timeout < 0")
	}
	return t, nil
}
