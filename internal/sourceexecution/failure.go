package sourceexecution

func rejected(request Request, sourceDigest, stage, code, message string) Receipt {
	return seal(Receipt{
		Schema: ReceiptSchema, Decision: "FAIL_CLOSED", Reason: code,
		Resolution: "EXACT", Filename: request.Filename, SourceDigest: sourceDigest,
		Events: []Event{}, Diagnostics: []Diagnostic{{Stage: stage, Code: code, Message: message}},
		Effects: Effects{},
	})
}
