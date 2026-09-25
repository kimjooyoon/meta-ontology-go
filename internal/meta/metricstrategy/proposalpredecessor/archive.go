package proposalpredecessor

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposal"
)

func decodeProposalArchive(payload []byte) ([]byte, proposal.Report, string, error) {
	if len(payload) == 0 || len(payload) > maximumResponseBytes {
		return nil, proposal.Report{}, "", &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: fmt.Errorf("proposal archive size is invalid")}
	}
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return nil, proposal.Report{}, "", &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: err}
	}
	proposalPayload, checksumPayload, proposalCount, checksumCount := []byte(nil), []byte(nil), 0, 0
	for _, file := range reader.File {
		switch path.Base(file.Name) {
		case "proposal-contract.json":
			proposalCount++
			proposalPayload, err = readArchiveFile(file)
		case "proposal-contract.json.sha256":
			checksumCount++
			checksumPayload, err = readArchiveFile(file)
		}
		if err != nil {
			return nil, proposal.Report{}, "", &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: err}
		}
	}
	if proposalCount != 1 || checksumCount != 1 {
		return nil, proposal.Report{}, "", &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: fmt.Errorf("proposal archive members are ambiguous")}
	}
	digestBytes := sha256.Sum256(proposalPayload)
	digest := hex.EncodeToString(digestBytes[:])
	fields := strings.Fields(string(checksumPayload))
	if len(fields) != 2 || fields[0] != digest || path.Base(fields[1]) != "proposal-contract.json" {
		return nil, proposal.Report{}, "", &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: fmt.Errorf("proposal checksum diverged")}
	}
	var report proposal.Report
	if err := json.Unmarshal(proposalPayload, &report); err != nil {
		return nil, proposal.Report{}, "", &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: err}
	}
	if err := proposal.Validate(report); err != nil {
		return nil, proposal.Report{}, "", &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: err}
	}
	return proposalPayload, report, "sha256:" + digest, nil
}
