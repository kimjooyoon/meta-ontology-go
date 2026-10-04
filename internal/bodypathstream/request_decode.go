package bodypathstream

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
)

func decodeStreamRequest(item record) (Request, error) {
	var request Request
	if err := validateStreamRecord(item); err != nil {
		return request, err
	}
	decoder := json.NewDecoder(bytes.NewReader(item.raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return request, errors.New("request has trailing JSON content")
		}
		return request, fmt.Errorf("request has trailing content: %w", err)
	}
	if request.Schema != RequestSchema {
		return request, fmt.Errorf("schema must be %q", RequestSchema)
	}
	if !validCorrelationID(request.CorrelationID) {
		return request, errors.New("correlation_id must be 1-128 UTF-8 bytes without control characters")
	}
	return request, nil
}

func validateStreamRecord(item record) error {
	if item.tooLarge || len(item.raw) > MaxRecordBytes {
		return fmt.Errorf("request exceeds %d bytes", MaxRecordBytes)
	}
	if len(item.raw) == 0 {
		return errors.New("empty NDJSON record")
	}
	if !utf8.Valid(item.raw) {
		return errors.New("request is not valid UTF-8")
	}
	if err := canonicalKeys(item.raw); err != nil {
		return err
	}
	if err := decision.RejectDuplicateJSONKeys(item.raw); err != nil {
		return fmt.Errorf("validate request JSON: %w", err)
	}
	return nil
}
