package proofchoicealgebra

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type directive struct {
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	ClaimID       string `json:"claim_id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Statement     string `json:"statement"`
	Choice        Choice `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Persistent    bool   `json:"persistent"`
}

func parseBundle(path string, source []byte) (Bundle, []issue) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || len(diagnostics) > 0 {
		return Bundle{}, []issue{{Reason: "SOURCE_PARSE_UNKNOWN"}}
	}
	bundle := Bundle{}
	var issues []issue
	for lineNumber, line := range strings.Split(string(source), "\n") {
		text := strings.TrimSpace(line)
		const prefix = "// proof-choice "
		if !strings.HasPrefix(text, prefix) {
			continue
		}
		var raw directive
		if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(text, prefix))), &raw); err != nil {
			issues = append(issues, issue{Reason: "DIRECTIVE_UNKNOWN", Line: lineNumber + 1})
			continue
		}
		switch strings.ToUpper(raw.Kind) {
		case "CLAIM", "METRIC":
			bundle.Items = append(bundle.Items, Item{
				Kind: Kind(strings.ToUpper(raw.Kind)), ID: raw.ID, Statement: raw.Statement,
				Choice: raw.Choice, Producer: raw.Producer, Consumer: raw.Consumer,
				MetaOperation: raw.MetaOperation, Stage: raw.Stage, Step: raw.Step,
				Reason: raw.Reason, Numerator: raw.Numerator, Denominator: raw.Denominator,
				Line: lineNumber + 1,
			})
		case "TRANSITION":
			bundle.Transitions = append(bundle.Transitions, Transition{
				ClaimID: raw.ClaimID, From: raw.From, To: raw.To, Choice: raw.Choice,
				Producer: raw.Producer, Consumer: raw.Consumer, MetaOperation: raw.MetaOperation,
				Stage: raw.Stage, Step: raw.Step, Reason: raw.Reason,
				Persistent: raw.Persistent, Line: lineNumber + 1,
			})
		default:
			issues = append(issues, issue{Reason: "DIRECTIVE_UNKNOWN", Line: lineNumber + 1})
		}
	}
	return bundle, issues
}

func digestSource(source []byte) string {
	sum := sha256.Sum256(source)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestReceipt(receipt Receipt) (string, error) {
	receipt.Digest = ""
	data, err := json.Marshal(receipt)
	if err != nil {
		return "", fmt.Errorf("marshal receipt for digest: %w", err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
