package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"
)

type ConfigurationDataApp struct {
	NODES []Node
}

type Node struct {
	Name               string
	ConnectionType     string
	IP                 string
	Port               int
	ID                 uint8
	Log                bool
	ConnectionTimeout  float64
	ConnectionAttempts uint
	TAGS               []Tag
}

type Tag struct {
	Name       string
	Address    uint32
	DataType   string
	ScanPeriod float64
}

func (tn *ConfigurationDataApp) Setup(c *ConfigHandler) error {
	content, err := os.ReadFile(c.fileName)
	if err != nil {
		return err
	}

	var tmpTN ConfigurationDataApp
	err = json.Unmarshal(content, &tmpTN)
	if err != nil {
		return err
	}

	// проверка полученных данных
	// цикл по узлам
	for i := 0; i < len(tmpTN.NODES); i++ {
		k := 0
		j := i + 1
		// считаем количество неодинаковых имен
		for ; j < len(tmpTN.NODES); j++ {
			if strings.TrimSpace(tmpTN.NODES[i].Name) != strings.TrimSpace(tmpTN.NODES[j].Name) {
				k++
			}
		}
		// если все имена неодинаковые, то проверяем полученные данные
		// и добавляем в выходной массив новый узел
		if (j - i - 1) == k {
			var tmpNode Node
			var err error

			tmpNode.Name, err = verifyName(tmpTN.NODES[i].Name)
			if err != nil {
				fmt.Println("App config: node skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNode.IP, err = verifyNodeIP(tmpTN.NODES[i].IP)
			if err != nil {
				fmt.Println("App config: Node skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}

			tmpNode.ConnectionType, err = verifyConnectionType(tmpTN.NODES[i].ConnectionType)
			if err != nil {
				fmt.Println("App config: Node skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}

			tmpNode.ID, err = verifyNodeID(tmpTN.NODES[i].ID)
			if err != nil {
				fmt.Println("App config: Node skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNode.Port, err = verifyNodePort(tmpTN.NODES[i].Port)
			if err != nil {
				fmt.Println("App config: Node skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNode.ConnectionTimeout, err = verifyNodeConnectionTimeout(tmpTN.NODES[i].ConnectionTimeout)
			if err != nil {
				fmt.Println("App config: Node skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			tmpNode.Log = tmpTN.NODES[i].Log
			tmpNode.ConnectionAttempts, err = verifyNodeConnectionAttempts(tmpTN.NODES[i].ConnectionAttempts)
			if err != nil {
				fmt.Println("App config: Node skipped Name: " + tmpTN.NODES[i].Name + " because " + err.Error())
				continue
			}
			// добавляем проверенный узел в массив NODES
			tn.NODES = append(tn.NODES, tmpNode)
		} else {
			// если параметры узла не прошли првоерку скипаем их
			fmt.Println("App config: Node skipped Name: " + tmpTN.NODES[i].Name + " because not unique identifier name")
			continue
		}

		// если итерируемый узел прошёл проверку проверяем его теги
		// цикл по тегам узла
		for i1 := 0; i1 < len(tmpTN.NODES[i].TAGS); i1++ {
			k := 0
			j := i1 + 1
			// считаем неодинаковые имена
			for ; j < len(tmpTN.NODES[i].TAGS); j++ {
				if tmpTN.NODES[i].TAGS[i1].Name != tmpTN.NODES[i].TAGS[j].Name {
					k++
				}
			}

			if (j - i1 - 1) == k {
				var tmpTag Tag
				var err error

				tmpTag.Name, err = verifyName(tmpTN.NODES[i].TAGS[i1].Name)
				if err != nil {
					fmt.Println("App config: Tag skipped Name: " + tmpTN.NODES[i].TAGS[i1].Name + " because " + err.Error())
					continue
				}
				tmpTag.DataType, err = verifyTagDataType(tmpTN.NODES[i].TAGS[i1].DataType)
				if err != nil {
					fmt.Println("App config: Tag skipped Name: " + tmpTN.NODES[i].TAGS[i1].Name + " because " + err.Error())
					continue
				}
				tmpTag.Address, err = verifyTagAddress(tmpTN.NODES[i].TAGS[i1].Address)
				if err != nil {
					fmt.Println("App config: Tag skipped Name: " + tmpTN.NODES[i].TAGS[i1].Name + " because " + err.Error())
					continue
				}

				tmpTag.ScanPeriod, err = verifyTagScanPeriod(tmpTN.NODES[i].TAGS[i1].ScanPeriod)
				if err != nil {
					fmt.Println("App config: Tag skipped Name: " + tmpTN.NODES[i].TAGS[i1].Name + " because " + err.Error())
					continue
				}
				// добавляем проверенные данные в последний узел
				tn.NODES[len(tn.NODES)-1].TAGS = append(tn.NODES[len(tn.NODES)-1].TAGS, tmpTag)
			}
		}
	}
	return nil
}

///функции выполняющие верификацию полученных данных  NODE -----{
func verifyName(name string) (string, error) { // эта функция также используется для верификации имени тэга
	rtn := strings.TrimSpace(name)

	if rtn == "" {
		return rtn, errors.New("config did not have Name")
	}
	return rtn, nil
}

func verifyConnectionType(ct string) (string, error) {
	rtn := strings.TrimSpace(ct)

	if rtn == "" {
		return rtn, errors.New("config did not have connection type")
	}
	return rtn, nil
}

func verifyNodeIP(ip string) (string, error) {
	ipAddr := net.ParseIP(strings.TrimSpace(ip))

	if ipAddr == nil {
		return strings.TrimSpace(ip), errors.New("config did not have Ip")
	}
	return strings.TrimSpace(ip), nil
}

func verifyNodePort(port int) (int, error) {
	if port == 0 {
		return port, errors.New("config did not have port")
	}
	return port, nil
}

func verifyNodeID(id uint8) (uint8, error) {
	if id == 0 {
		return id, errors.New("config did not have ID")
	}
	return id, nil
}

func verifyNodeConnectionTimeout(ct float64) (float64, error) {
	if ct < 0.001 {
		return ct, errors.New("config connection timeout < 0.001")
	}
	return ct, nil
}

func verifyNodeConnectionAttempts(ca uint) (uint, error) {
	if ca < 1 {
		return ca, errors.New("config connection attempts < 1")
	}
	return ca, nil
}

///функции выполняющте верификацию полученных данных NODE -----}

///функции выполняющте верификацию полученных данных для тега TAG -----{
func verifyTagAddress(address uint32) (uint32, error) {
	return address - 1, nil
}

func verifyTagDataType(dataType string) (string, error) { // эта функция также используется для верификации имени тэга
	str := strings.TrimSpace(strings.ToLower(dataType))

	switch str {
	case constants.COIL_TYPE:
		return str, nil
	case constants.DWORD_TYPE:
		return str, nil
	case constants.WORD_TYPE:
		return str, nil
	default:
		return str, errors.New("config did not have type")
	}
}

func verifyTagScanPeriod(sp float64) (float64, error) {
	if sp < 0.001 {
		return sp, errors.New("config scan period < 0.001")
	}
	return sp, nil
}

// }
