package win

import (
	"fmt"
	"sync/atomic"
	"time"

	configuration "zheleznovux.com/modbus-console/cmd/configuration"
	tags "zheleznovux.com/modbus-console/cmd/storage"
	"zheleznovux.com/modbus-console/cmd/win/commander"
)

type WinConfig struct {
	nodeCommand []configuration.NodeTag
}

type WinConfigMgr struct {
	config atomic.Value
}

var winConfigMgr = &WinConfigMgr{}
var changeCh chan int = make(chan int)

func (a *WinConfigMgr) Callback(conf *configuration.ConfigHandler) {
	winConfig := &WinConfig{}
	winConfig.nodeCommand = conf.GetConfig().(*configuration.ConfigurationDataWin).NODES
	winConfigMgr.config.Store(winConfig)
	changeCh <- 1
}

func InitConfig(file string) {
	conf, err := configuration.NewConfig(file)
	if err != nil {
		fmt.Printf("read config file err: %v\n", err)
		return
	}

	conf.AddObserver(winConfigMgr)

	var winConfig WinConfig
	winConfig.nodeCommand = conf.GetConfig().(*configuration.ConfigurationDataWin).NODES

	winConfigMgr.config.Store(&winConfig)
	fmt.Println("Выполнена загрузка конфигурации команд")
}

func Run(th *tags.TagsHandler) {
	winConfig := winConfigMgr.config.Load().(*WinConfig)
	var channelCount int
	quit := make(chan int)

	fmt.Println("Запущен обработчик")
	for {
		select {
		case <-changeCh:
			{
				for j := 0; j < channelCount; j++ {
					quit <- 1
				}
				channelCount = 0

				for i := range winConfig.nodeCommand {
					c, err := commander.Setup(winConfig.nodeCommand[i], th)
					if err != nil {
						fmt.Println(err)
						continue
					}

					go c.StartChecking(quit)
					channelCount++
				}
			}
		default:
			{
				time.Sleep(2 * time.Second)
			}
		}
	}
}
