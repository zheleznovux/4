package app

import (
	"fmt"
	"sync/atomic"
	"time"

	modbus "zheleznovux.com/modbus-console/cmd/app/modbus"
	configuration "zheleznovux.com/modbus-console/cmd/configuration"
	tags "zheleznovux.com/modbus-console/cmd/storage"
)

type AppConfig struct {
	clients []modbus.Client
}

type AppConfigMgr struct {
	config atomic.Value
}

var appConfigMgr = &AppConfigMgr{}
var changeCh chan int

func (a *AppConfigMgr) Callback(conf *configuration.ConfigHandler) {
	appConfig := &AppConfig{}
	appConfig.clients = setup(conf)
	fmt.Println(appConfig.clients)

	changeCh <- 1
	appConfigMgr.config.Store(appConfig)
}

func setup(ch *configuration.ConfigHandler) []modbus.Client {

	config := ch.GetConfig()
	rtn := make([]modbus.Client, 0)

	var err error
	nodes := config.(*configuration.ConfigurationDataNode).NODES
	for i := range nodes {
		var tmp modbus.Client
		tmp, err = modbus.NewClinet(nodes[i].IP, nodes[i].Port, nodes[i].ID, nodes[i].Name)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for j := range nodes[i].TAGS {
			err := tmp.SetTag(
				nodes[i].TAGS[j].Name,
				nodes[i].TAGS[j].Address,
				nodes[i].TAGS[j].ScanPeriod,
				nodes[i].TAGS[j].DataType,
				nodes[i].TAGS[j].DataBit)

			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		rtn = append(rtn, tmp)
	}

	return rtn
}

func InitConfig(file string, th *tags.TagsHandler) {
	conf, err := configuration.NewConfig(file)
	if err != nil {
		fmt.Printf("read config file err: %v\n", err)
		return
	}

	conf.AddObserver(appConfigMgr)
	conf.AddObserver(th)

	var appConfig AppConfig
	appConfig.clients = setup(conf)
	th.SetData(tags.Setup(conf))

	appConfigMgr.config.Store(&appConfig)
	fmt.Println("Выполнена загрузка конфигурации тэгов")
}

func startSender(client modbus.Client, clientId int, tagId int, quit chan int, th *tags.TagsHandler) {
	fmt.Println("Запущен опрос тега " + client.Name() + "." + client.Tags()[tagId].GetName())

	for {
		select {
		case <-quit:
			{
				return
			}
		default:
			{
				err := client.Send(tagId)
				if err != nil {
					fmt.Println(err)
				} else {
					th.SetDataTag(clientId, tagId, &client.Tags()[tagId])
				}
				time.Sleep(time.Duration(client.Tags()[tagId].GetScanPeriod()) * time.Second)

			}
		}

	}
}

// func tryConnect() {
// 	for {
// 		err := appConfig.clients[clientId].Connect()
// 		if err != nil {
// 			continue
// 		}
// 	}

// }

func Run(th *tags.TagsHandler) {
	var channelCount int
	appConfig := appConfigMgr.config.Load().(*AppConfig)

	quit := make(chan int)
	changeCh = make(chan int)

	for {
		select {
		case <-changeCh:
			{
				fmt.Println("app change")

				for j := 0; j < channelCount; j++ {
					quit <- 0
				}
				channelCount = 0

				for clientId := range appConfig.clients {
					err := appConfig.clients[clientId].Connect()
					if err != nil {
						fmt.Println(err)
					}
					for tagId := range appConfig.clients[clientId].Tags() {
						go startSender(appConfig.clients[clientId], clientId, tagId, quit, th)
						channelCount++
					}
				}
			}
		default:
			time.Sleep(2 * time.Second)
			th.Save()
		}

	}

}
