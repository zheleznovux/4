package constants

// опреаторы сравнения
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

// логические операторы
const (
	AND = "and"
	OR  = "or"
)

// команды для win
const (
	SHUTDOWN    = "shutdown"
	RESTART     = "restart"
	RUN_PROGRAM = "run"
)

// типы тэгов
const (
	WORD_TYPE  = "word"
	COIL_TYPE  = "coil"
	DWORD_TYPE = "dword"
)

// состояния сети
const (
	GOOD = "good"
	BAD  = "bad"
)

// протоколы
const (
	MODBUS_TCP = "modbusTCP"
)

// функции
const (
	FUNCTION_1 = 0x1
	FUNCTION_2 = 0x2
	FUNCTION_3 = 0x3
	FUNCTION_4 = 0x4
	FUNCTION_5 = 0x5
	FUNCTION_6 = 0x6
)

const (
	UINT8_MAX_VALUE  = 0xFF
	UINT16_MAX_VALUE = 0xFFFF
	UINT32_MAX_VALUE = 0xFFFFFFFF
)

// var OPERATOR = [...]string{
// 	EQUAL,
// 	NOT_EQUAL,
// 	MORE_EQUAL,
// 	LESS_EQUAL,
// 	MORE,
// 	LESS}
