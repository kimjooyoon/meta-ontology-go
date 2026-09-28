package retentionpolicy

import (
	"errors"
	"fmt"
	"go/format"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	generatedpolicy "github.com/kimjooyoon/meta-ontology-go/internal/meta/retentionpolicy/generated"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	EvaluatorSchema       = "gooo/semantic-retention-policy-evaluator/v1"
	PolicyActivity        = "PublishOperationReceipt"
	RetentionDecisionHead = "retention-decision:v1"

	DecisionClosed  = "CLOSED"
	DecisionUnknown = "UNKNOWN"
	DecisionRefuted = "REFUTED"

	CaseCertificateHit        = "CERTIFICATE_HIT"
	CaseCertificateReplay     = "CERTIFICATE_REPLAY"
	CaseMissingCertificate    = "MISSING_CERTIFICATE"
	CaseMissingAuthorization  = "MISSING_AUTHORIZATION"
	CaseStaleInput            = "STALE_INPUT"
	CaseMismatchedCertificate = "MISMATCHED_CERTIFICATE"
)

var caseIDs = [...]string{
	"CERTIFICATE_HIT",
	"CERTIFICATE_REPLAY",
	"MISSING_CERTIFICATE",
	"MISSING_AUTHORIZATION",
	"STALE_INPUT",
	"MISMATCHED_CERTIFICATE",
}

// DecisionCase is the pure decision row lowered from the .gooo authority.
type DecisionCase struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

// Policy is the source-derived retained-knowledge decision policy. The
// generated evaluator is the executable form, while Cases remains the source
// meaning that the loader checks against it before the compiler can use it.
type Policy struct {
	SourceDigest    string
	EvaluatorDigest string
	ActivityCount   int
	Cases           []DecisionCase
}

func CaseIDs() []string { return append([]string(nil), caseIDs[:]...) }

func GeneratedEvaluatorDigest() string { return generatedpolicy.EvaluatorDigest }

func GeneratedEvaluatorCaseIDs() []string { return generatedpolicy.CaseIDs() }

func Evaluate(caseID string) (string, bool) { return generatedpolicy.Evaluate(caseID) }

func (policy Policy) Decision(caseID string) (string, bool) {
	return generatedpolicy.Evaluate(caseID)
}

// CompileNamed lowers the authoritative .gooo source and extracts the pure
// retained decision rows from the PublishOperationReceipt activity.
func CompileNamed(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("retention policy source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Policy{}, fmt.Errorf("parse retention policy: %w", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower retention policy: %w", err)
	}
	if ir.Package != "selfimprovementobservation" || ir.Namespace.String() != "self_improvement_observation" {
		return Policy{}, fmt.Errorf("retention policy package/namespace is %q/%q", ir.Package, ir.Namespace)
	}
	var rows []DecisionCase
	activityCount := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != PolicyActivity {
			continue
		}
		activityCount++
		var err error
		rows, err = parseRows(node.ValueProgram)
		if err != nil {
			return Policy{}, fmt.Errorf("activity %q: %w", PolicyActivity, err)
		}
	}
	if activityCount != 1 {
		return Policy{}, fmt.Errorf("retention policy activity count = %d, want 1", activityCount)
	}
	return Policy{
		SourceDigest:    cache.HashBytes(source).String(),
		EvaluatorDigest: evaluatorDigest(rows),
		ActivityCount:   activityCount,
		Cases:           rows,
	}, nil
}

func GenerateNamed(filename string, source []byte) (Policy, []byte, error) {
	policy, err := CompileNamed(filename, source)
	if err != nil {
		return Policy{}, nil, err
	}
	generated, err := render(filename, policy)
	if err != nil {
		return Policy{}, nil, err
	}
	return policy, generated, nil
}

func Load(filename string, source []byte) (Policy, error) {
	policy, err := CompileNamed(filename, source)
	if err != nil {
		return Policy{}, err
	}
	if policy.EvaluatorDigest != generatedpolicy.EvaluatorDigest {
		return Policy{}, errors.New("retention policy and generated evaluator digest differ")
	}
	if generatedpolicy.PolicySourceDigest != policy.SourceDigest {
		return Policy{}, errors.New("retention policy source digest differs from generated evaluator")
	}
	if len(generatedpolicy.CaseIDs()) != len(policy.Cases) {
		return Policy{}, errors.New("retention policy and generated evaluator case counts differ")
	}
	for index, row := range policy.Cases {
		if generatedpolicy.CaseIDs()[index] != row.ID {
			return Policy{}, fmt.Errorf("retention evaluator case %q is not bound to source row %d", row.ID, index+1)
		}
		decision, ok := generatedpolicy.Evaluate(row.ID)
		if !ok || decision != row.Decision {
			return Policy{}, fmt.Errorf("retention evaluator decision for %q is not source-bound", row.ID)
		}
	}
	return policy, nil
}

func parseRows(value string) ([]DecisionCase, error) {
	parts := strings.Split(value, ";")
	seenHead := false
	rows := make([]DecisionCase, 0, len(caseIDs))
	seen := make(map[string]bool, len(caseIDs))
	for _, part := range parts {
		switch {
		case part == RetentionDecisionHead:
			if seenHead {
				return nil, errors.New("retention decision header is duplicated")
			}
			seenHead = true
		case strings.HasPrefix(part, "retention-case="):
			encoded := strings.TrimPrefix(part, "retention-case=")
			id, decision, ok := strings.Cut(encoded, ":")
			if !ok || id == "" || decision == "" || strings.Contains(decision, ":") || seen[id] {
				return nil, fmt.Errorf("retention decision row %q is malformed or duplicated", encoded)
			}
			if !knownCase(id) || !knownDecision(decision) {
				return nil, fmt.Errorf("retention decision row %q is outside the fixed policy", encoded)
			}
			seen[id] = true
			rows = append(rows, DecisionCase{ID: id, Decision: decision})
		}
	}
	if !seenHead || len(rows) != len(caseIDs) {
		return nil, fmt.Errorf("retention decision rows = %d, want %d", len(rows), len(caseIDs))
	}
	for index, id := range caseIDs {
		if rows[index].ID != id {
			return nil, fmt.Errorf("retention decision row %d is %q, want %q", index+1, rows[index].ID, id)
		}
	}
	return rows, nil
}

func knownCase(id string) bool {
	for _, known := range caseIDs {
		if id == known {
			return true
		}
	}
	return false
}

func knownDecision(decision string) bool {
	return decision == DecisionClosed || decision == DecisionUnknown || decision == DecisionRefuted
}

func evaluatorDigest(rows []DecisionCase) string {
	var builder strings.Builder
	builder.WriteString(EvaluatorSchema)
	builder.WriteByte('\n')
	for _, row := range rows {
		builder.WriteString(row.ID)
		builder.WriteByte('=')
		builder.WriteString(row.Decision)
		builder.WriteByte('\n')
	}
	return cache.HashBytes([]byte(builder.String())).String()
}

func render(filename string, policy Policy) ([]byte, error) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "// Code generated by gooo retention policy generator; DO NOT EDIT.\n\n")
	fmt.Fprintf(&builder, "package retentionpolicygenerated\n\n")
	fmt.Fprintf(&builder, "const (\n")
	fmt.Fprintf(&builder, "\tEvaluatorSchema = %q\n", EvaluatorSchema)
	fmt.Fprintf(&builder, "\tPolicySourcePath = %q\n", filename)
	fmt.Fprintf(&builder, "\tPolicySourceDigest = %q\n", policy.SourceDigest)
	fmt.Fprintf(&builder, "\tEvaluatorDigest = %q\n", policy.EvaluatorDigest)
	fmt.Fprintf(&builder, ")\n\n")
	fmt.Fprintf(&builder, "type DecisionCase struct { ID string; Decision string }\n\n")
	fmt.Fprintf(&builder, "var cases = [...]DecisionCase{\n")
	for _, row := range policy.Cases {
		fmt.Fprintf(&builder, "\t{ID: %q, Decision: %q},\n", row.ID, row.Decision)
	}
	fmt.Fprintf(&builder, "}\n\n")
	fmt.Fprintf(&builder, "func CaseIDs() []string {\n\tresult := make([]string, 0, len(cases))\n\tfor _, row := range cases { result = append(result, row.ID) }\n\treturn result\n}\n\n")
	fmt.Fprintf(&builder, "func Evaluate(caseID string) (string, bool) {\n\tswitch caseID {\n")
	for _, decision := range []string{DecisionClosed, DecisionUnknown, DecisionRefuted} {
		ids := make([]string, 0, len(policy.Cases))
		for _, row := range policy.Cases {
			if row.Decision == decision {
				ids = append(ids, row.ID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		fmt.Fprintf(&builder, "\tcase ")
		for index, id := range ids {
			if index != 0 {
				builder.WriteString(", ")
			}
			fmt.Fprintf(&builder, "%q", id)
		}
		fmt.Fprintf(&builder, ":\n\t\treturn %q, true\n", decision)
	}
	fmt.Fprintf(&builder, "\tdefault:\n\t\treturn \"\", false\n\t}\n}\n")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return nil, fmt.Errorf("format generated retention evaluator: %w", err)
	}
	return formatted, nil
}
