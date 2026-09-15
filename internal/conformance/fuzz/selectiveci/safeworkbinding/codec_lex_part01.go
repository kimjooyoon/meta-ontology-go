package safeworkbinding

type jsonValueKind uint8

const (
	jsonNullValue   jsonValueKind = 0
	jsonBoolValue   jsonValueKind = 1
	jsonNumberValue jsonValueKind = 2
	jsonStringValue jsonValueKind = 3
	jsonArrayValue  jsonValueKind = 4
	jsonObjectValue jsonValueKind = 5
)

type jsonValue struct {
	kind   jsonValueKind
	text   string
	object map[string]jsonValue
}
type jsonParser struct {
	data      []byte
	offset    int
	duplicate bool
}

func hexValue(value byte) (byte, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10, true
	case value >= 'A' && value <= 'F':
		return value - 'A' + 10, true
	default:
		return 0, false
	}
}
