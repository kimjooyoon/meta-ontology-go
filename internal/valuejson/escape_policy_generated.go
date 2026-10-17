package valuejson

//gooo:generated:start id="urn:gooo:text-input:escape-policy" kind="activity"
func TextEscapePolicy(input0 int64, input1 int64) string {

	var unit = input0
	var next = input1
	if unit < 0 || unit > 65535 {
		return "INVALID_UNIT"
	}
	if unit >= 55296 && unit <= 56319 {
		if next >= 56320 && next <= 57343 {
			return "SURROGATE_PAIR"
		}
		return "UNPAIRED_HIGH"
	}
	if unit >= 56320 && unit <= 57343 {
		return "UNPAIRED_LOW"
	}
	return "SCALAR_ESCAPE"

}

//gooo:generated:end id="urn:gooo:text-input:escape-policy" kind="activity"
