package win

import (
	"fmt"
	"sync"
	"sync/atomic"

	configuration "zheleznovux.com/modbus-console/cmd/configuration"
	server "zheleznovux.com/modbus-console/cmd/serverStorage"
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
}

func Run(th *server.Server) {
	winConfig := winConfigMgr.config.Load().(*WinConfig)
	quit := make(chan struct{})
	var wg sync.WaitGroup

	for {
		<-changeCh
		close(quit)
		wg.Wait()
		quit = make(chan struct{})

		for i := range winConfig.nodeCommand {
			c, err := commander.Setup(winConfig.nodeCommand[i], th)
			if err != nil {
				fmt.Println(err)
				continue
			}
			wg.Add(1)
			go c.StartChecking(quit, &wg)
		}

	}

}
