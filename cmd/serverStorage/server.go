package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"

	"zheleznovux.com/modbus-console/cmd/serverStorage/client"
	"zheleznovux.com/modbus-console/cmd/serverStorage/tag"
)

// var _ configuration.Notifyer = (nil)

type Server struct {
	data   []client.ClientInterface
	rwLock sync.RWMutex
	Sync   bool
}

var changeCh chan int = make(chan int)

func (thisSH *Server) Callback(conf *configuration.ConfigHandler) {
	thisSH.Setup(conf)
	fmt.Println(thisSH.data[0])
	changeCh <- 1
}

func New() *Server {
	return &Server{}
}

func (thisSH *Server) Setup(confHandler *configuration.ConfigHandler) {
	config := confHandler.GetConfig()
	rtn := make([]client.ClientInterface, 0)

	nodes := config.(*configuration.ConfigurationDataApp).NODES
	for i := range nodes {
		var tmp client.ClientInterface
		switch nodes[i].ConnectionType {
		case constants.MODBUS_TCP:
			{
				var err error

				tmp, err = client.NewClinetModbus(nodes[i].IP, nodes[i].Port, nodes[i].ID, nodes[i].Name, nodes[i].Debug, int(nodes[i].ConnectionAttempts), nodes[i].ConnectionTimeout)
				if err != nil {
					fmt.Println(err)
					continue
				}
				for j := range nodes[i].TAGS {
					t, err := tag.NewTag(
						nodes[i].TAGS[j].Name,
						nodes[i].TAGS[j].Address,
						nodes[i].TAGS[j].ScanPeriod,
						nodes[i].TAGS[j].DataType)
					if err != nil {
						fmt.Println(err.Error())
						continue
					}
					err = tmp.SetTag(t)
					if err != nil {
						fmt.Println(err.Error())
						continue
					}
				}
			}
		default:
			{
				fmt.Println("неизвестный тип подключения")
				continue
			}
		}
		rtn = append(rtn, tmp)
	}
	thisSH.rwLock.RLock()
	defer thisSH.rwLock.RUnlock()
	thisSH.data = rtn
}

func (thisSH *Server) GetData() []client.ClientInterface {
	thisSH.rwLock.RLock()
	defer thisSH.rwLock.RUnlock()
	return thisSH.data
}

func (thisSH *Server) GetTagByName(name string) (tag.TagInterface, error) {

	split := strings.Split(name, ".")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid name")
	}

	thisSH.rwLock.RLock()
	defer thisSH.rwLock.RUnlock()

	for i := range thisSH.data {
		if thisSH.data[i].Name() == split[0] {
			for j := range thisSH.data[i].Tags() {
				if thisSH.data[i].Tags()[j].Name() == split[1] {
					return thisSH.data[i].Tags()[j], nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no such name")
}

func (thisSH *Server) Save() {
	thisSH.rwLock.RLock()
	defer thisSH.rwLock.RUnlock()

	rankingsJson, err := json.Marshal(thisSH.data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = ioutil.WriteFile("output.json", rankingsJson, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (thisSH *Server) Run() {
	quit := make(chan struct{})
	var wg sync.WaitGroup

	for {
		// сигнал смены конфига
		select {
		case <-changeCh:
			{
				close(quit)
				wg.Wait()
				quit = make(chan struct{})
				fmt.Println("app change")

				for clientId := range thisSH.data {
					wg.Add(1)
					go thisSH.data[clientId].Start(quit, &wg)
				}
			}
		default:
			time.Sleep(10 * time.Second)
			thisSH.Save()
		}
	}
}
