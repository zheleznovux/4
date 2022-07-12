package tags

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type BaseTag interface {
	SetName(string)
	GetName() string

	SetAddress(string) error
	GetAddress() uint16

	SetScanPeriod(float64) error
	GetScanPeriod() float64

	SetValue(uint16)
	GetValue() uint16

	SetDataType(string)
	GetDataType() string

	SetState(bool)
	GetState() bool

	SetTimestamp()
	GetTimestamo() string

	ReadSending(value uint16)
}

type CoilTag struct {
	Name       string
	DataType   string
	Address    uint16
	ScanPeriod float64
	Bit        uint8
	Value      uint16
	Timestamp  string
	State      bool
}

type WordTag struct {
	Name       string
	DataType   string
	Address    uint16
	ScanPeriod float64
	Value      uint16
	Timestamp  string
	State      bool
}

func (t *CoilTag) SetTimestamp() {
	now := time.Now()
	t.Timestamp = now.Format(time.RFC3339)
}
func (t *CoilTag) GetTimestamo() string {
	return t.Timestamp
}

func (t *WordTag) SetTimestamp() {
	now := time.Now()
	t.Timestamp = now.Format(time.RFC3339)
}
func (t *WordTag) GetTimestamo() string {
	return t.Timestamp
}

func (t *CoilTag) SetState(state bool) {
	t.State = state
}
func (t *CoilTag) GetState() bool {
	return t.State
}

func (t *WordTag) SetState(state bool) {
	t.State = state
}
func (t *WordTag) GetState() bool {
	return t.State
}

func (t *CoilTag) SetDataType(dataType string) {
	t.DataType = dataType
}
func (t *CoilTag) GetDataType() string {
	return t.DataType
}

func (t *WordTag) SetDataType(dataType string) {
	t.DataType = dataType
}
func (t *WordTag) GetDataType() string {
	return t.DataType
}

func (t *CoilTag) ReadSending(value uint16) {
	mask := uint16(math.Pow(2, float64(t.Bit)))
	if (value & mask) == mask {
		t.SetValue(1)
	} else {
		t.SetValue(0)
	}
}

func (t *WordTag) ReadSending(value uint16) {
	t.SetValue(value)
}

func (t *CoilTag) SetValue(value uint16) {
	t.SetTimestamp()
	t.SetState(true)
	t.Value = value
}

func (t *WordTag) SetValue(value uint16) {
	t.SetTimestamp()
	t.SetState(true)
	t.Value = value
}

func (t *CoilTag) GetValue() uint16 {
	return t.Value
}

func (t *WordTag) GetValue() uint16 {
	return t.Value
}

func (t *CoilTag) SetName(name string) {
	t.Name = strings.TrimSpace(name)
}

func (t *WordTag) SetName(name string) {
	t.Name = strings.TrimSpace(name)
}

func (t *CoilTag) GetName() string {
	return t.Name
}

func (t *WordTag) GetName() string {
	return t.Name
}

func (t *CoilTag) SetAddress(address string) error {
	address = strings.TrimSpace(address)

	tmp, err := strconv.Atoi(address)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if tmp < 0 {
		return fmt.Errorf("address < 0 ")
	}
	t.Address = uint16(tmp)
	return nil
}

func (t *CoilTag) GetAddress() uint16 {
	return t.Address
}
func (t *WordTag) GetAddress() uint16 {
	return t.Address
}

func (t *WordTag) SetAddress(address string) error {
	address = strings.TrimSpace(address)

	tmp, err := strconv.Atoi(address)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if tmp < 0 {
		return fmt.Errorf("address < 0 ")
	}
	t.Address = uint16(tmp)
	return nil
}

func (t *CoilTag) SetScanPeriod(time float64) error {
	if time < 0 {
		return fmt.Errorf("time < 0")
	}
	t.ScanPeriod = time
	return nil
}

func (t *CoilTag) GetScanPeriod() float64 {
	return t.ScanPeriod
}

func (t *WordTag) GetScanPeriod() float64 {
	return t.ScanPeriod
}

func (t *WordTag) SetScanPeriod(time float64) error {
	if time < 0 {
		return fmt.Errorf("time < 0")
	}
	t.ScanPeriod = time
	return nil
}

func (t *CoilTag) SetBit(bit uint8) error {
	if bit > 15 {
		return fmt.Errorf("(bit < 0) || (bit > 15)")
	}
	t.Bit = bit
	return nil
}
