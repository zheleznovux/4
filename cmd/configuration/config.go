package configuration

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type ConfigurationData interface{}

type ConfigHandler struct {
	data           ConfigurationData
	fileName       string
	lastModifyTime int64
	rwLock         sync.RWMutex
	notifyList     []Notifyer
}

func (c *ConfigHandler) parse() (ConfigurationData, error) {
	var tmpСonf ConfigurationData
	if strings.Contains(c.fileName, "win_") {
		tmpСonf = &ConfigurationDataTagNode{}
		tmpСonf.(*ConfigurationDataTagNode).Setup(c)
	} else {
		tmpСonf = &ConfigurationDataNode{}
		tmpСonf.(*ConfigurationDataNode).Setup(c)
	}

	return tmpСonf, nil
}

func NewConfig(fileName string) (conf *ConfigHandler, err error) {

	conf = &ConfigHandler{
		fileName: fileName,
	}

	m, err := conf.parse()
	if err != nil {
		fmt.Printf("parse conf error:%v\n", err)
		return
	}

	conf.rwLock.Lock()
	conf.data = m
	conf.rwLock.Unlock()

	go conf.reload()
	return conf, nil
}

func (c *ConfigHandler) GetConfig() ConfigurationData {
	c.rwLock.RLock()
	defer c.rwLock.RUnlock()
	return c.data
}

func (c *ConfigHandler) reload() {
	ticker := time.NewTicker(time.Second * 5)

	for _ = range ticker.C {

		func() {
			f, err := os.Open(c.fileName)
			if err != nil {
				fmt.Printf("reload: open file error:%s\n", err)
				return
			}
			defer f.Close()

			fileInfo, err := f.Stat()
			if err != nil {
				fmt.Printf("stat file error:%s\n", err)
				return
			}

			curModifyTime := fileInfo.ModTime().Unix()
			if curModifyTime > c.lastModifyTime {
				m, err := c.parse()
				if err != nil {
					fmt.Printf("parse config error:%v\n", err)
					return
				}

				c.rwLock.Lock()
				c.data = m
				c.rwLock.Unlock()

				c.lastModifyTime = curModifyTime

				for _, n := range c.notifyList {
					n.Callback(c)
				}
			}
		}()
	}
}

func (c *ConfigHandler) AddObserver(n Notifyer) {
	c.notifyList = append(c.notifyList, n)
}
