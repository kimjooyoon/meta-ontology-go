package verify

func Verify(input Input, closureJSON []byte) (Report, error) {
	input, err := normalizeInput(input)
	if err != nil {
		return Report{}, err
	}
	if err := validateIdentity(input); err != nil {
		return Report{}, err
	}
	actual, program, verification, err := decodeInput(input, closureJSON)
	if err != nil {
		return Report{}, err
	}
	if err := validateDocuments(input, program, verification); err != nil {
		return Report{}, err
	}
	expected, err := reconstruct(input, program, verification)
	if err != nil {
		return Report{}, err
	}
	if err := compareReceipts(actual, expected); err != nil {
		return Report{}, err
	}
	return newReport(actual)
}
