package constants

// все опреаторы сравнения, читаемые программой
const (
	EQUAL      = "=="
	NOT_EQUAL  = "!="
	MORE_EQUAL = ">="
	LESS_EQUAL = "<="
	MORE       = ">"
	LESS       = "<"
	BIT        = "bit"
	NOT_BIT    = "!bit"
)

// все логические операторы, читаемые программой
const (
	AND = "and"
	OR  = "or"
)

// команды, читаемые программой
const (
	SHUTDOWN    = "shutdown"
	RESTART     = "restart"
	RUN_PROGRAM = "run"
)

// типы, читаемые программой
const (
	WORD_TYPE  = "word"
	COIL_TYPE  = "coil"
	DWORD_TYPE = "dword"
)

// состояния сети, читаемые программой
const (
	GOOD = "good"
	BAD  = "bad"
)

// протоколы, читаемые программой
const (
	MODBUS_TCP = "modbusTCP"
)

// var OPERATOR = [...]string{
// 	EQUAL,
// 	NOT_EQUAL,
// 	MORE_EQUAL,
// 	LESS_EQUAL,
// 	MORE,
// 	LESS}
