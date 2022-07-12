package app

import (
	"fmt"
	"sync/atomic"
	"time"

	"zheleznovux.com/modbus-console/cmd/app/modbus"
	"zheleznovux.com/modbus-console/cmd/configuration"
	"zheleznovux.com/modbus-console/cmd/tags"
)

type AppConfig struct {
	clients []modbus.Client
}

type AppConfigMgr struct {
	config atomic.Value
}

var appConfigMgr = &AppConfigMgr{}
var changeCh chan int = make(chan int)

func (a *AppConfigMgr) Callback(conf *configuration.ConfigHandler) {
	appConfig := &AppConfig{}
	appConfig.clients = Setup(conf)
	changeCh <- 1
	appConfigMgr.config.Store(appConfig)
}

func Setup(ch *configuration.ConfigHandler) []modbus.Client {
	config := ch.GetConfig()
	rtn := make([]modbus.Client, 0)
	var err error

	for i := range config.NODES {
		var tmp modbus.Client
		tmp, err = modbus.NewClinet(config.NODES[i].IP, config.NODES[i].Port, config.NODES[i].ID, config.NODES[i].Name)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for j := range config.NODES[i].TAGS {
			err := tmp.SetTag(
				config.NODES[i].TAGS[j].Name,
				config.NODES[i].TAGS[j].Address,
				config.NODES[i].TAGS[j].ScanPeriod,
				config.NODES[i].TAGS[j].DataType,
				config.NODES[i].TAGS[j].DataBit)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		rtn = append(rtn, tmp)
	}

	return rtn
}

func InitConfig(file string, ts *tags.TagsHandler) {
	conf, err := configuration.NewConfig(file)
	if err != nil {
		fmt.Printf("read config file err: %v\n", err)
		return
	}

	conf.AddObserver(appConfigMgr)
	conf.AddObserver(ts)

	var appConfig AppConfig
	appConfig.clients = Setup(conf)
	ts.SetData(tags.Setup(conf))

	appConfigMgr.config.Store(&appConfig)
	fmt.Println("Выполнена загрузка конфигурации")
}

func startSender(client modbus.Client, clientId int, tagId int, quit chan int, ts *tags.TagsHandler) {
	for {
		select {
		default:
			{
				err := client.Send(tagId)
				if err != nil {
					fmt.Println(err)
				} else {
					ts.SetDataTag(clientId, tagId, &client.GetTags()[tagId])
				}
				time.Sleep(time.Duration(client.GetTags()[tagId].GetScanPeriod()) * time.Second)

			}
		case <-quit:
			{
				return

			}
		}

	}
}

// context

func Run(ts *tags.TagsHandler) {
	fmt.Println("Запущено")
	var channelCount int
	appConfig := appConfigMgr.config.Load().(*AppConfig)
	quit := make(chan int)
	for {
		select {
		case <-changeCh:
			{
				for j := 0; j < channelCount; j++ {
					quit <- 0
				}
				channelCount = 0

				for clientId := range appConfig.clients {
					err := appConfig.clients[clientId].Connect()
					if err != nil {
						fmt.Println(err)
					}
					for tagId := range appConfig.clients[clientId].GetTags() {
						go startSender(appConfig.clients[clientId], clientId, tagId, quit, ts)
						channelCount++
					}
				}
			}
		default:
			time.Sleep(10 * time.Second)
			ts.Save()
		}

	}

}
