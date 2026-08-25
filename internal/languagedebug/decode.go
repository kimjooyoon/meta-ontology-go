package languagedebug

import "encoding/json"

type executionReceipt struct {
	Schema         string            `json:"schema"`
	Decision       string            `json:"decision"`
	Resolution     string            `json:"resolution"`
	Filename       string            `json:"filename"`
	SourceDigest   string            `json:"source_digest"`
	SemanticDigest string            `json:"semantic_digest"`
	Entry          json.RawMessage   `json:"entry"`
	Events         []Event           `json:"events"`
	Diagnostics    []json.RawMessage `json:"diagnostics"`
	Effects        Effects           `json:"effects"`
	Digest         string            `json:"digest"`
}

func decodeExecution(data []byte) (executionReceipt, bool) {
	var receipt executionReceipt
	if json.Unmarshal(data, &receipt) != nil {
		return executionReceipt{}, false
	}
	valid := receipt.Schema == "gooo/source-execution-receipt/v1" &&
		receipt.Decision == "PASS" && receipt.Resolution == ResolutionExact &&
		receipt.Filename != "" && validDigest(receipt.SourceDigest) &&
		validDigest(receipt.SemanticDigest) && validDigest(receipt.Digest) &&
		len(receipt.Events) > 0
	return receipt, valid
}
