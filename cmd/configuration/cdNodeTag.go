package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigurationDataTagNode struct {
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

func (tn *ConfigurationDataTagNode) Setup(c *ConfigHandler) error {
	content, err := os.ReadFile(c.fileName)
	if err != nil {
		return err
	}

	var tmpTN ConfigurationDataTagNode
	err = json.Unmarshal(content, &tmpTN)
	if err != nil {
		return err
	}

	for i := 0; i < len(tmpTN.NODES); i++ {
		k := 0
		j := i + 1
		for ; j < len(tmpTN.NODES); j++ {
			if tmpTN.NODES[i].Name != tmpTN.NODES[j].Name {
				k++
			}
		}

		if (j - i - 1) == k {
			err := checkSpaceNodeTag(tmpTN.NODES[i])
			if err != nil {
				fmt.Println(err)
				continue
			}
			tn.NODES = append(tn.NODES, tmpTN.NODES[i])
		}
	}

	return nil
}

func checkSpaceNodeTag(nt NodeTag) error {
	switch "" {
	case nt.Name:
		return fmt.Errorf("config did not have Name")
	case nt.DataType:
		return fmt.Errorf("config did not have Datatype")
	case nt.StateCondition:
		return fmt.Errorf("config did not have StateCondition")
	case nt.ValueCondition:
		return fmt.Errorf("config did not have ValueCondition")
	case nt.Logic:
		return fmt.Errorf("config did not have Logic")
	case nt.Action:
		return fmt.Errorf("config did not have Action")
	default:
		break
	}

	if nt.ScanPeriod < 0.0001 {
		return fmt.Errorf("config did not have ScanPeriod")
	}

	return nil
}
