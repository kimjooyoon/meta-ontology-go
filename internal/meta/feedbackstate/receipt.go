package feedbackstate

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type archivedReceipt struct {
	Schema              string         `json:"schema"`
	Report              archivedReport `json:"report"`
	ReplayReportDigest  string         `json:"replay_report_digest"`
	ReplayVerified      bool           `json:"replay_verified"`
	RepositoryWrites    int            `json:"repository_writes"`
	ReceiptDigest       string         `json:"receipt_digest"`
}

type archivedReport struct {
	Schema              string           `json:"schema"`
	Feedback            archivedFeedback `json:"feedback"`
	SourceDecision      string           `json:"source_decision"`
	Decision            string           `json:"decision"`
	Reason              string           `json:"reason"`
	FromResolution      string           `json:"from_resolution"`
	ToResolution        string           `json:"to_resolution"`
	PreviousDescents    int              `json:"previous_descents"`
	Descents            int              `json:"descents"`
	RepositoryWrites    int              `json:"repository_writes"`
	ReportDigest        string           `json:"report_digest"`
}

type archivedFeedback struct {
	CommitSHA      string `json:"commit_sha"`
	Repository     string `json:"repository"`
	Decision       string `json:"decision"`
	NextOperation  string `json:"next_operation"`
}

func decode(raw []byte) (archivedReceipt, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var receipt archivedReceipt
	if err := decoder.Decode(&receipt); err != nil { return receipt, err }
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return archivedReceipt{}, errors.New("receipt has trailing JSON")
	}
	return receipt, nil
}

func payloadDigest(raw []byte) string {
	return bytesDigest(raw)
}
