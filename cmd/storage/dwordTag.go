package tags

import (
	"encoding/json"
	"errors"
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

func (dwt *DWordTag) Setup(name string, address uint16, scanPeriod float64) error {
	var err error
	err = dwt.SetName(name)
	if err != nil {
		return err
	}
	err = dwt.SetAddress(address)
	if err != nil {
		return err
	}
	dwt.SetDataType()
	err = dwt.SetScanPeriod(scanPeriod)
	if err != nil {
		return err
	}
	dwt.SetState(false)
	return nil
}

func (dwt DWordTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name      string
		DataType  string
		Address   uint16
		Value     uint32
		Timestamp string
		State     bool
	}{
		Name:      dwt.name,
		DataType:  dwt.dataType,
		Address:   dwt.address,
		Value:     dwt.value,
		Timestamp: dwt.timestamp,
		State:     dwt.state,
	})
}

//===================================Name
func (dwt *DWordTag) SetName(name string) error {
	tmp := strings.TrimSpace(name)
	if tmp == "" {
		return errors.New("empty tag name")
	}
	dwt.name = tmp
	return nil
}
func (t *DWordTag) Name() string {
	return t.name
}

//===================================DataType
func (dwt *DWordTag) SetDataType() {
	dwt.dataType = constants.DWORD_TYPE
}
func (dwt *DWordTag) DataType() string {
	return dwt.dataType
}

//===================================Address
func (dwt *DWordTag) Address() uint16 {
	return dwt.address
}
func (dwt *DWordTag) SetAddress(address uint16) error {
	if address == 0xFF {
		return errors.New("address == 0xFF")
	}
	dwt.address = address
	return nil
}

//===================================TimeStamp
func (dwt *DWordTag) SetTimestamp() {
	now := time.Now()
	dwt.timestamp = now.Format(time.RFC3339)
}
func (dwt *DWordTag) Timestamp() string {
	return dwt.timestamp
}

//===================================State
func (dwt *DWordTag) SetState(state bool) {
	dwt.state = state
}
func (dwt *DWordTag) State() bool {
	return dwt.state
}

//===================================Value не интерфейсный метод
func (dwt *DWordTag) SetValue(value uint32) {
	dwt.SetTimestamp()
	dwt.SetState(true)
	dwt.value = value
}
func (dwt *DWordTag) Value() uint32 {
	return dwt.value
}

//===================================ScanPeriod
func (dwt *DWordTag) ScanPeriod() float64 {
	return dwt.scanPeriod
}
func (dwt *DWordTag) SetScanPeriod(time float64) error {
	if time < 0 {
		return fmt.Errorf("time < 0")
	}
	dwt.scanPeriod = time
	return nil
}
