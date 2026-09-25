package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"time"
)

func writeProvenanceFailure(writer io.Writer, response provenancePublishResponse, code, message string) int {
	response.Status = provenanceStatusRejected
	response.StoreDigest = ""
	response.Records = []provenanceCLIRecord{}
	response.Error = &provenanceCLIError{Code: code, Message: message}
	if err := sealProvenanceResponse(&response); err != nil {
		return exitFailure
	}
	if err := writeProvenanceJSON(writer, response); err != nil {
		return exitFailure
	}
	return exitFailure
}
func sealProvenanceResponse(response *provenancePublishResponse) error {
	response.CanonicalDigest = ""
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	response.CanonicalDigest = sha256Digest(payload)
	return nil
}
func writeProvenanceResponse(writer io.Writer, response provenancePublishResponse, deadline time.Time) error {
	if err := sealProvenanceResponse(&response); err != nil {
		return err
	}
	return writeProvenanceJSONWithDeadline(writer, response, deadline)
}
func writeProvenanceJSON(writer io.Writer, response provenancePublishResponse) error {
	return writeProvenanceJSONWithDeadline(writer, response, time.Time{})
}
func writeProvenanceJSONWithDeadline(writer io.Writer, response provenancePublishResponse, deadline time.Time) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if len(payload) > maxDiagnosticBytes {
		return errDiagnosticLimit
	}
	return writeInspectOutput(writer, payload, deadline)
}
func sha256Digest(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
