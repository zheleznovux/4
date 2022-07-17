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
var changeCh chan int = make(chan int)

func (a *AppConfigMgr) Callback(conf *configuration.ConfigHandler) {
	appConfig := &AppConfig{}
	appConfig.clients = setup(conf)
	appConfigMgr.config.Store(appConfig)

	changeCh <- 1
}

// инициализация клиентов из конфигурационного файла
func setup(ch *configuration.ConfigHandler) []modbus.Client {
	config := ch.GetConfig()
	rtn := make([]modbus.Client, 0)

	var err error
	nodes := config.(*configuration.ConfigurationDataApp).NODES
	for i := range nodes {
		var tmp modbus.Client
		tmp, err = modbus.NewClinet(nodes[i].IP, nodes[i].Port, nodes[i].ID, nodes[i].Name, nodes[i].Debug, int(nodes[i].ConnectionAttempts), nodes[i].ConnectionTimeout)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for j := range nodes[i].TAGS {
			err := tmp.SetTag(
				nodes[i].TAGS[j].Name,
				nodes[i].TAGS[j].Address,
				nodes[i].TAGS[j].ScanPeriod,
				nodes[i].TAGS[j].DataType)

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

func startSender(client modbus.Client, clientId int, tagId int, quit chan int, counter *int, th *tags.TagsHandler) {
	fmt.Println("Запущен опрос тега " + client.Name() + "." + client.Tags()[tagId].Name())
	for {
		select {
		case <-quit:
			{
				*counter--
				fmt.Println("Завершен опрос тега " + client.Name() + "." + client.Tags()[tagId].Name())
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
				// fmt.Printf("%s.%s = %d\n", client.Name(), client.Tags()[tagId].Name(), client.Tags()[tagId].Value())
				time.Sleep(time.Duration(client.Tags()[tagId].ScanPeriod()) * time.Second)
			}
		}

	}
}

func Run(th *tags.TagsHandler) {
	var channelCount int

	quit := make(chan int)

	for {
		// сигнал смены конфига
		select {
		case <-changeCh:
			{
				fmt.Println("app change")

				appConfig := appConfigMgr.config.Load().(*AppConfig)

				for j := 0; j != channelCount; {
					quit <- 0
				}
				channelCount = 0

				for clientId := range appConfig.clients {
					go func(clientId int) {
						for i := 0; i < appConfig.clients[clientId].ConnectionAttempts(); i++ {
							err := appConfig.clients[clientId].Connect()
							if err != nil {
								fmt.Println(err.Error() + " в узле " + appConfig.clients[clientId].Name())
							} else {
								for tagId := range appConfig.clients[clientId].Tags() {
									go startSender(appConfig.clients[clientId], clientId, tagId, quit, &channelCount, th)
									channelCount++
								}
								return
							}
							time.Sleep(appConfig.clients[clientId].ConnectionTimeout())
						}
						fmt.Println("Достигнуто максимальное количество попыток подключения в узле " + appConfig.clients[clientId].Name())
					}(clientId)
				}
			}
		default:
			time.Sleep(2 * time.Second)
			th.Save()
		}

	}
}
