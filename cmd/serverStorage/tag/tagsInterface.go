package tag

import (
	"errors"

	"zheleznovux.com/modbus-console/cmd/serverStorage/constants"
)

type TagInterface interface {
	SetName(string) error
	Name() string

	SetAddress(uint32) error
	Address() uint16

	SetScanPeriod(float64) error
	ScanPeriod() float64

	SetDataType()
	DataType() string

	SetState(bool)
	State() bool

	SetTimestamp()
	Timestamp() string

	Setup(name string, address uint32, scanPeriod float64) error
}

func NewTag(name string, address uint32, scanPeriod float64, dataType string) (TagInterface, error) {
	var tagI TagInterface
	if dataType == "" {
		return nil, errors.New("invalid dataType {constructor tag}")
	}
	switch dataType {
	case constants.COIL_TYPE:
		{
			var tag CoilTag
			tagI = &tag
		}
	case constants.WORD_TYPE:
		{
			var tag WordTag
			tagI = &tag
		}
	case constants.DWORD_TYPE:
		{
			var tag DWordTag
			tagI = &tag
		}
	default:
		return nil, errors.New("invalid dataType {constructor tag}")
	}

	err := tagI.Setup(name, address, scanPeriod)
	if err != nil {
		return nil, err
	}
	return tagI, nil
}
