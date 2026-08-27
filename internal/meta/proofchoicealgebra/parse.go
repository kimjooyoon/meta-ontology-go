package proofchoicealgebra

import (
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

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
		appendDirective(&bundle, &issues, raw, lineNumber+1)
	}
	return bundle, issues
}

func appendDirective(bundle *Bundle, issues *[]issue, raw directive, line int) {
	switch strings.ToUpper(raw.Kind) {
	case "CLAIM", "METRIC":
		bundle.Items = append(bundle.Items, Item{Kind: Kind(strings.ToUpper(raw.Kind)), ID: raw.ID, Statement: raw.Statement, Choice: raw.Choice, Producer: raw.Producer, Consumer: raw.Consumer, MetaOperation: raw.MetaOperation, Stage: raw.Stage, Step: raw.Step, Reason: raw.Reason, Numerator: raw.Numerator, Denominator: raw.Denominator, Line: line})
	case "TRANSITION":
		bundle.Transitions = append(bundle.Transitions, Transition{ClaimID: raw.ClaimID, From: raw.From, To: raw.To, Choice: raw.Choice, Producer: raw.Producer, Consumer: raw.Consumer, MetaOperation: raw.MetaOperation, Stage: raw.Stage, Step: raw.Step, Reason: raw.Reason, Persistent: raw.Persistent, Line: line})
	default:
		*issues = append(*issues, issue{Reason: "DIRECTIVE_UNKNOWN", Line: line})
	}
}
