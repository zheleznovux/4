package commander

type Condition interface {
	checkValue()
}

type dwordCondition struct {
	operator       string
	valueCondition uint32
}
