package tags

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/constants"
)

type CoilTag struct { //не могу заприватить из-за json парса
	name       string
	dataType   string
	address    uint16
	scanPeriod float64
	value      byte
	timestamp  string
	state      bool
}

func (ct CoilTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name      string
		DataType  string
		Address   uint16
		Value     byte
		Timestamp string
		State     bool
	}{
		Name:      ct.name,
		DataType:  ct.dataType,
		Address:   ct.address,
		Value:     ct.value,
		Timestamp: ct.timestamp,
		State:     ct.state,
	})
}

//===================================Name
func (t *CoilTag) SetName(name string) {
	t.name = strings.TrimSpace(name)
}
func (t *CoilTag) Name() string {
	return t.name
}

//===================================DataType
func (t *CoilTag) SetDataType() {
	t.dataType = constants.COIL_TYPE
}
func (t *CoilTag) DataType() string {
	return t.dataType
}

//===================================Address
func (t *CoilTag) SetAddress(address uint16) {
	t.address = address
}
func (t *CoilTag) Address() uint16 {
	return t.address
}

//===================================ScanPeriod
func (t *CoilTag) SetScanPeriod(time float64) error {
	if time < 0 {
		return fmt.Errorf("time < 0")
	}
	t.scanPeriod = time
	return nil
}
func (t *CoilTag) ScanPeriod() float64 {
	return t.scanPeriod
}

//===================================Value
func (t *CoilTag) SetValue(value byte) {
	t.SetTimestamp()
	t.SetState(true)
	t.value = value
}
func (t *CoilTag) Value() byte {
	return t.value
}

//===================================TimeStamp
func (t *CoilTag) SetTimestamp() {
	now := time.Now()
	t.timestamp = now.Format(time.RFC3339)
}
func (t *CoilTag) Timestamp() string {
	return t.timestamp
}

//===================================DataState
func (t *CoilTag) SetState(state bool) {
	t.state = state
}
func (t *CoilTag) State() bool {
	return t.state
}
