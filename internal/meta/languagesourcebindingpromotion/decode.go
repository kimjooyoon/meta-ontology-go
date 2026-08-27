package languagesourcebindingpromotion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func decodeStrict[T any](raw []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return value, fmt.Errorf("trailing JSON input")
	}
	return value, nil
}

func decodeView[T any](raw []byte) (T, error) {
	var value T
	err := json.Unmarshal(raw, &value)
	return value, err
}

func verifyProducerDigest(value producerEnvelope) bool {
	digest := value.Digest
	value.Digest = ""
	return validDigest(digest) && digest == digestJSON(value)
}

func verifyReceiptDigest(value receiptEnvelope) bool {
	digest := value.Digest
	value.Digest = ""
	return validDigest(digest) && digest == digestJSON(value)
}

func verifyOracleDigest(value oracleEnvelope) bool {
	digest := value.Digest
	value.Digest = ""
	return validDigest(digest) && digest == digestJSON(value)
}
