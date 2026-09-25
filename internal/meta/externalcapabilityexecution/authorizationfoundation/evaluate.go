package authorizationfoundation

func Evaluate(input Input) (Receipt, error) {
	foundation, metadata, prior, current, err := decodeInput(input)
	if err != nil {
		return Receipt{}, err
	}
	if err := validateFoundation(foundation); err != nil {
		return Receipt{}, err
	}
	if err := validateMetadata(metadata, foundation); err != nil {
		return Receipt{}, err
	}
	if digestRaw(input.PriorReceiptRaw) != foundation.ReceiptFileDigest {
		return Receipt{}, denied("POLICY_FOUNDATION_RECEIPT_FILE_MISMATCH")
	}
	if err := validateBootstrap(prior, foundation.SubjectSHA, foundation, true); err != nil {
		return Receipt{}, err
	}
	if err := validateBootstrap(current, input.ExpectedSubject, foundation, false); err != nil {
		return Receipt{}, err
	}
	result := closeReceipt(current, foundation)
	if err := Validate(result, input.ExpectedSubject); err != nil {
		return Receipt{}, err
	}
	return result, nil
}
