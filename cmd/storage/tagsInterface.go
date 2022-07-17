package tags

type TagInterface interface {
	SetName(string)
	Name() string

	SetAddress(uint16)
	Address() uint16

	SetScanPeriod(float64) error
	ScanPeriod() float64

	// SetValue(uint16)
	// Value() uint16

	SetDataType()
	DataType() string

	SetState(bool)
	State() bool

	SetTimestamp()
	Timestamp() string
}
