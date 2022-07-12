package main

import (
	"os"
	"strings"

	"zheleznovux.com/modbus-console/cmd/app"
	"zheleznovux.com/modbus-console/cmd/tags"
)

func main() {
	ts := tags.New()

	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
		if cmd == "sync" {
			ts.Sync = true
		}
	}

	app.InitConfig("config.json", ts)
	app.Run(ts)
}
