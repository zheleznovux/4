package tags

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"zheleznovux.com/modbus-console/cmd/configuration"
)

var _ configuration.Notifyer = (nil)

type TagsHandler struct {
	data   []StateNode
	rwLock sync.RWMutex
	Sync   bool
}

type StateNode struct {
	Name string
	Tags []BaseTag
}

func New() *TagsHandler {
	return &TagsHandler{}
}

func Setup(conf *configuration.ConfigHandler) []StateNode {
	t := make([]StateNode, 0)
	for i := range conf.GetConfig().NODES {
		var sn StateNode
		sn.Name = conf.GetConfig().NODES[i].Name
		for j := range conf.GetConfig().NODES[i].TAGS {
			switch conf.GetConfig().NODES[i].TAGS[j].DataType {
			case "coil":
				{
					var bt CoilTag
					bt.SetName(conf.GetConfig().NODES[i].TAGS[j].Name)
					bt.SetAddress(conf.GetConfig().NODES[i].TAGS[j].Address)
					bt.SetDataType(conf.GetConfig().NODES[i].TAGS[j].DataType)
					bt.SetScanPeriod(conf.GetConfig().NODES[i].TAGS[j].ScanPeriod)
					bt.SetBit(conf.GetConfig().NODES[i].TAGS[j].DataBit)
					bt.SetState(false)
					sn.Tags = append(sn.Tags, &bt)
				}
			default:
				{
					var bt WordTag
					bt.SetName(conf.GetConfig().NODES[i].TAGS[j].Name)
					bt.SetAddress(conf.GetConfig().NODES[i].TAGS[j].Address)
					bt.SetDataType(conf.GetConfig().NODES[i].TAGS[j].DataType)
					bt.SetScanPeriod(conf.GetConfig().NODES[i].TAGS[j].ScanPeriod)
					bt.SetState(false)
					sn.Tags = append(sn.Tags, &bt)
				}
			}

		}
		t = append(t, sn)
	}

	return t
}

func (ts *TagsHandler) SetData(sn []StateNode) {
	ts.rwLock.RLock()
	defer ts.rwLock.RUnlock()
	ts.data = sn
}

func (ts *TagsHandler) GetData() []StateNode {
	ts.rwLock.RLock()
	defer ts.rwLock.RUnlock()
	return ts.data
}

func (ts *TagsHandler) SetDataTag(clientId int, tagId int, tag *BaseTag) {
	ts.rwLock.RLock()
	defer ts.rwLock.RUnlock()
	ts.data[clientId].Tags[tagId] = *tag
	if ts.Sync {
		go ts.Save()
	}
}

func (ts *TagsHandler) Callback(conf *configuration.ConfigHandler) {
	ts.SetData(Setup(conf))
}

func (ts *TagsHandler) Save() {
	ts.rwLock.RLock()
	defer ts.rwLock.RUnlock()

	rankingsJson, err := json.Marshal(ts.data)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	err = ioutil.WriteFile("output.json", rankingsJson, 0644)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
}
