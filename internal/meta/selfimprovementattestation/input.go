package selfimprovementattestation

import (
	"bytes"
	"encoding/json"
	"os"
)

func LoadRequest(receiptPath, archivePath, verificationPath string, exitCode int, version string) (Request, error) {
	var request Request
	receiptData, err := os.ReadFile(receiptPath)
	if err != nil {
		return request, err
	}
	if err := json.Unmarshal(receiptData, &request.TransportReceipt); err != nil {
		return request, err
	}
	request.ArchiveDigest, request.ArchiveProducer, request.ArchiveObservationDigest, err = readArchive(archivePath)
	if err != nil {
		return request, err
	}
	verificationData, err := os.ReadFile(verificationPath)
	if err != nil {
		return request, err
	}
	if len(bytes.TrimSpace(verificationData)) > 0 {
		if err := json.Unmarshal(verificationData, &request.Verification); err != nil {
			return request, err
		}
	}
	request.VerifierExitCode = exitCode
	request.VerifierVersion = version
	return request, nil
}
