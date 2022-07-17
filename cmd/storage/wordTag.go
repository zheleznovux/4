package tags

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/constants"
)

type WordTag struct {
	name       string
	dataType   string
	address    uint16
	scanPeriod float64
	value      uint16
	timestamp  string
	state      bool
}

func (wt WordTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name      string
		DataType  string
		Address   uint16
		Value     uint16
		Timestamp string
		State     bool
	}{
		Name:      wt.name,
		DataType:  wt.dataType,
		Address:   wt.address,
		Value:     wt.value,
		Timestamp: wt.timestamp,
		State:     wt.state,
	})
}

//===================================Name
func (t *WordTag) SetName(name string) {
	t.name = strings.TrimSpace(name)
}
func (t *WordTag) Name() string {
	return t.name
}

//===================================DataType
func (t *WordTag) SetDataType() {
	t.dataType = constants.WORD_TYPE
}
func (t *WordTag) DataType() string {
	return t.dataType
}

//===================================Address
func (t *WordTag) Address() uint16 {
	return t.address
}
func (t *WordTag) SetAddress(address uint16) {
	t.address = address
}

//===================================TimeStamp
func (t *WordTag) SetTimestamp() {
	now := time.Now()
	t.timestamp = now.Format(time.RFC3339)
}
func (t *WordTag) Timestamp() string {
	return t.timestamp
}

//===================================State
func (t *WordTag) SetState(state bool) {
	t.state = state
}
func (t *WordTag) State() bool {
	return t.state
}

//===================================Value не интерфейсный метод
func (t *WordTag) SetValue(value uint16) {
	t.SetTimestamp()
	t.SetState(true)
	t.value = value
}
func (t *WordTag) Value() uint16 {
	return t.value
}

//===================================ScanPeriod
func (t *WordTag) ScanPeriod() float64 {
	return t.scanPeriod
}
func (t *WordTag) SetScanPeriod(time float64) error {
	if time < 0 {
		return fmt.Errorf("time < 0")
	}
	t.scanPeriod = time
	return nil
}
