package closure

import "fmt"

func Build(input Input) (Receipt, error) {
	input, err := normalizeInput(input)
	if err != nil {
		return Receipt{}, err
	}
	program, verification, err := decodeDocuments(input)
	if err != nil {
		return Receipt{}, err
	}
	if err := validateProgram(input, program); err != nil {
		return Receipt{}, err
	}
	if digestBytes(input.Source) != program.SourceDigest {
		return Receipt{}, fmt.Errorf("program source digest does not match program.gooo")
	}
	if err := validateVerification(input, program, verification); err != nil {
		return Receipt{}, err
	}
	receipt := newReceipt(input, program, verification)
	receipt.Indicators, err = buildIndicators(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.Digest, err = digestReceipt(receipt)
	return receipt, err
}
