package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/configuration"
	storage "zheleznovux.com/modbus-console/cmd/serverStorage"
)

func InitConfig(file string, storageHandler *storage.Server) {
	conf, err := configuration.NewConfig(file)
	if err != nil {
		fmt.Printf("read config file err: %v\n", err)
		return
	}

	conf.AddObserver(storageHandler)
	storageHandler.Setup(conf)

	fmt.Println("Выполнена загрузка конфигурации тэгов")
}

func main() {
	storageHandler := storage.New()

	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		if cmd == "sync" {
			storageHandler.Sync = true
		}
	}
	InitConfig("config.json", storageHandler)
	// win.InitConfig("win_config.json")

	go storageHandler.Run()

	// go win.Run(storageHandler)

	for {
		time.Sleep(200000000)
	}
}
