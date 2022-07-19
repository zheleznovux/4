package tag

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"
)

type CoilTag struct {
	name       string
	dataType   string
	address    uint16
	scanPeriod float64
	value      byte
	timestamp  string
	state      bool
}

func (ct *CoilTag) Setup(name string, address uint32, scanPeriod float64) error {
	var err error
	err = ct.SetName(name)
	if err != nil {
		return err
	}
	err = ct.SetAddress(address)
	if err != nil {
		return err
	}
	ct.SetDataType()
	err = ct.SetScanPeriod(scanPeriod)
	if err != nil {
		return err
	}
	ct.SetState(false)
	return nil
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
func (t *CoilTag) SetName(name string) error {
	tmp := strings.TrimSpace(name)
	if tmp == "" {
		return errors.New("empty tag name")
	}
	t.name = tmp
	return nil
}
func (t CoilTag) Name() string {
	return t.name
}

//===================================DataType
func (t *CoilTag) SetDataType() {
	t.dataType = constants.COIL_TYPE
}
func (t CoilTag) DataType() string {
	return t.dataType
}

//===================================Address
func (t *CoilTag) SetAddress(address uint32) error {
	if address >= constants.UINT16_MAX_VALUE {
		tmpINT := int(address / 100000.0)
		if (tmpINT != constants.FUNCTION_1) && (tmpINT != constants.FUNCTION_2) {
			return errors.New("invalid function address")
		}
		tmpUINT16 := uint16(address - uint32(tmpINT*100000))
		if tmpUINT16 >= constants.UINT16_MAX_VALUE {
			return errors.New("invalid tag address")
		}
		t.address = tmpUINT16
		return nil
	}
	t.address = uint16(address)
	return nil
}
func (t CoilTag) Address() uint16 {
	return t.address
}

//===================================ScanPeriod
func (t *CoilTag) SetScanPeriod(time float64) error {
	if time < 0 {
		return errors.New("time < 0")
	}
	t.scanPeriod = time
	return nil
}
func (t CoilTag) ScanPeriod() float64 {
	return t.scanPeriod
}

//===================================Value
func (t *CoilTag) SetValue(value byte) {
	t.SetTimestamp()
	t.SetState(true)
	t.value = value
}
func (t CoilTag) Value() byte {
	return t.value
}

//===================================TimeStamp
func (t *CoilTag) SetTimestamp() {
	now := time.Now()
	t.timestamp = now.Format(time.RFC3339)
}
func (t CoilTag) Timestamp() string {
	return t.timestamp
}

//===================================DataState
func (t *CoilTag) SetState(state bool) {
	t.state = state
}
func (t CoilTag) State() bool {
	return t.state
}
