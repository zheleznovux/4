package commander

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger struct {
	Lock           sync.Mutex
	ParentNodeName string
	IsLogOutput    bool
}

const (
	INFO    = "INFO"
	WARNING = "WARNING"
	ERROR   = "ERROR"
)

func (thisLog *Logger) Write(level string, msg string) {

	text := fmt.Sprintf("%s >> Node.Tag: name - %s. Message [%s]: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		thisLog.ParentNodeName,
		level,
		msg)

	fmt.Print(text)
	if thisLog.IsLogOutput {
		thisLog.Lock.Lock()
		defer thisLog.Lock.Unlock()
		save(text)
	}
}

func save(text string) {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		fmt.Println(err)
	}
}
