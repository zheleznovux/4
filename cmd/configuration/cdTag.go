package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigurationDataNode struct {
	NODES []Node
}

type Node struct {
	Name string
	IP   string
	Port int
	ID   uint8
	TAGS []Tag
}

type Tag struct {
	Name       string
	Address    string
	DataType   string
	ScanPeriod float64
	DataBit    uint8
}

func (tn *ConfigurationDataNode) Setup(c *ConfigHandler) error {
	content, err := os.ReadFile(c.fileName)
	if err != nil {
		return err
	}

	var tmpTN ConfigurationDataNode
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
			var tmp Node
			tmp.ID = tmpTN.NODES[i].ID
			tmp.IP = tmpTN.NODES[i].IP
			tmp.Name = tmpTN.NODES[i].Name
			tmp.Port = tmpTN.NODES[i].Port
			err := checkSpaceNode(tmp)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			tn.NODES = append(tn.NODES, tmp)
		} else {
			continue
		}

		for i1 := 0; i1 < len(tmpTN.NODES[i].TAGS); i1++ {
			k := 0
			j := i1 + 1
			for ; j < len(tmpTN.NODES[i].TAGS); j++ {
				if tmpTN.NODES[i].TAGS[i1].Name != tmpTN.NODES[i].TAGS[j].Name {
					k++
				}
			}

			if (j - i1 - 1) == k {
				err := checkSpaceTag(tmpTN.NODES[i].TAGS[i1])
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				tn.NODES[len(tn.NODES)-1].TAGS = append(tn.NODES[len(tn.NODES)-1].TAGS, tmpTN.NODES[i].TAGS[i1])
			}
		}
	}
	return nil
}

func checkSpaceNode(nt Node) error {
	switch "" {
	case nt.Name:
		return fmt.Errorf("Config did not have Name")
	case nt.IP:
		return fmt.Errorf("Config did not have IP")
	default:
		break
	}

	if nt.ID == 0 {
		return fmt.Errorf("Config did not have ID")
	} else if nt.Port < 0 {
		return fmt.Errorf("Config did not have Port")
	}

	return nil
}

func checkSpaceTag(nt Tag) error {
	switch "" {
	case nt.Name:
		return fmt.Errorf("Config did not have Name")
	case nt.Address:
		return fmt.Errorf("Config did not have Address")
	case nt.DataType:
		return fmt.Errorf("Config did not have DataType")
	default:
		break
	}

	if nt.ScanPeriod < 0.0001 {
		return fmt.Errorf("Config did not have ScanPeriod")
	} else if nt.DataBit > 1 {
		return fmt.Errorf("Config did not have DataBit")
	}

	return nil
}
