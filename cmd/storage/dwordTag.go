package tags

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/constants"
)

type DWordTag struct {
	name       string
	dataType   string
	address    uint16
	scanPeriod float64
	value      uint32
	timestamp  string
	state      bool
}

func (wc DWordTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name      string
		DataType  string
		Address   uint16
		Value     uint32
		Timestamp string
		State     bool
	}{
		Name:      wc.name,
		DataType:  wc.dataType,
		Address:   wc.address,
		Value:     wc.value,
		Timestamp: wc.timestamp,
		State:     wc.state,
	})
}

//===================================Name
func (t *DWordTag) SetName(name string) {
	t.name = strings.TrimSpace(name)
}
func (t *DWordTag) Name() string {
	return t.name
}

//===================================DataType
func (t *DWordTag) SetDataType() {
	t.dataType = constants.DWORD_TYPE
}
func (t *DWordTag) DataType() string {
	return t.dataType
}

//===================================Address
func (t *DWordTag) Address() uint16 {
	return t.address
}
func (t *DWordTag) SetAddress(address uint16) {
	t.address = address
}

//===================================TimeStamp
func (t *DWordTag) SetTimestamp() {
	now := time.Now()
	t.timestamp = now.Format(time.RFC3339)
}
func (t *DWordTag) Timestamp() string {
	return t.timestamp
}

//===================================State
func (t *DWordTag) SetState(state bool) {
	t.state = state
}
func (t *DWordTag) State() bool {
	return t.state
}

//===================================Value не интерфейсный метод
func (t *DWordTag) SetValue(value uint32) {
	t.SetTimestamp()
	t.SetState(true)
	t.value = value
}
func (t *DWordTag) Value() uint32 {
	return t.value
}

//===================================ScanPeriod
func (t *DWordTag) ScanPeriod() float64 {
	return t.scanPeriod
}
func (t *DWordTag) SetScanPeriod(time float64) error {
	if time < 0 {
		return fmt.Errorf("time < 0")
	}
	t.scanPeriod = time
	return nil
}
