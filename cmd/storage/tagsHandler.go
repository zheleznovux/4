package tags

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"zheleznovux.com/modbus-console/cmd/configuration"
	"zheleznovux.com/modbus-console/cmd/constants"
)

var _ configuration.Notifyer = (nil)

type TagsHandler struct {
	data   []StateNode
	rwLock sync.RWMutex
	Sync   bool
}

type StateNode struct {
	Name string
	Tags []TagInterface
}

func New() *TagsHandler {
	return &TagsHandler{}
}

func Setup(confHandler *configuration.ConfigHandler) []StateNode {
	t := make([]StateNode, 0)
	conf := confHandler.GetConfig().(*configuration.ConfigurationDataApp)
	for i := range conf.NODES {
		var sn StateNode
		sn.Name = conf.NODES[i].Name
		for j := range conf.NODES[i].TAGS {
			switch conf.NODES[i].TAGS[j].DataType {
			case constants.COIL_TYPE: //завернуть в коснтурктор!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
				{
					var bt CoilTag
					bt.SetName(conf.NODES[i].TAGS[j].Name)
					bt.SetAddress(conf.NODES[i].TAGS[j].Address)
					bt.SetDataType()
					bt.SetScanPeriod(conf.NODES[i].TAGS[j].ScanPeriod)
					bt.SetState(false)
					sn.Tags = append(sn.Tags, &bt)
				}
			case constants.WORD_TYPE:
				{
					var bt WordTag
					bt.SetName(conf.NODES[i].TAGS[j].Name)
					bt.SetAddress(conf.NODES[i].TAGS[j].Address)
					bt.SetDataType()
					bt.SetScanPeriod(conf.NODES[i].TAGS[j].ScanPeriod)
					bt.SetState(false)
					sn.Tags = append(sn.Tags, &bt)
				}
			case constants.DWORD_TYPE:
				{
					var bt DWordTag
					bt.SetName(conf.NODES[i].TAGS[j].Name)
					bt.SetAddress(conf.NODES[i].TAGS[j].Address)
					bt.SetDataType()
					bt.SetScanPeriod(conf.NODES[i].TAGS[j].ScanPeriod)
					bt.SetState(false)
					sn.Tags = append(sn.Tags, &bt)
				}
			default:
				continue
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

func (ts *TagsHandler) GetTagByName(name string) (TagInterface, error) {

	split := strings.Split(name, ".")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid name")
	}

	ts.rwLock.RLock()
	defer ts.rwLock.RUnlock()

	for i := range ts.data {
		if ts.data[i].Name == split[0] {
			for j := range ts.data[i].Tags {
				if ts.data[i].Tags[j].Name() == split[1] {
					return ts.data[i].Tags[j], nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no such name")
}

func (ts *TagsHandler) SetDataTag(clientId int, tagId int, tag *TagInterface) {
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
		fmt.Println(err.Error())
		return
	}
	err = ioutil.WriteFile("output.json", rankingsJson, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
