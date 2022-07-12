package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Node struct {
	Name string
	IP   string
	Port int
	ID   uint8
	TAGS []Tag
}

type Tag struct {
	Name       string
	Address    string
	DataType   string
	ScanPeriod float64
	DataBit    uint8
}

type ConfigurationData struct {
	NODES []Node
}

type ConfigHandler struct {
	data           ConfigurationData
	fileName       string
	lastModifyTime int64
	rwLock         sync.RWMutex
	notifyList     []Notifyer
}

func (c *ConfigHandler) parse() (ConfigurationData, error) {
	conf := ConfigurationData{}

	content, err := os.ReadFile(c.fileName)

	if err != nil {
		fmt.Println(err)
		return conf, err
	}

	json.Unmarshal(content, &conf)

	return conf, nil
}

func NewConfig(file string) (conf *ConfigHandler, err error) {
	conf = &ConfigHandler{
		fileName: file,
	}

	m, err := conf.parse()
	if err != nil {
		fmt.Println("parse conf error:%v\n", err)
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
