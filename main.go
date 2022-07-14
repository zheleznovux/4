package main

import (
	"os"
	"runtime"
	"strings"
	"time"

	app "zheleznovux.com/modbus-console/cmd/app"
	tags "zheleznovux.com/modbus-console/cmd/storage"
	win "zheleznovux.com/modbus-console/cmd/win"
)

func main() {
	runtime.GOMAXPROCS(4)

	ts := tags.New()

	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		if cmd == "sync" {
			ts.Sync = true
		}
	}
	app.InitConfig("config.json", ts)
	win.InitConfig("win_config.json")

	go app.Run(ts)
	go win.Run(ts)

	for {
		time.Sleep(200000000)
	}
}
