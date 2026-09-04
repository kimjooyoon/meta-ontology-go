package discoverypolicy

import (
	"errors"
	"fmt"
	"go/format"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	generatedpolicy "github.com/kimjooyoon/meta-ontology-go/internal/meta/discoverypolicy/generated"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	EvaluatorSchema         = "gooo/public-self-observation-policy-evaluator/v1"
	PolicyActivity          = "ClassifyPublicObservation"
	DiscoveryDecisionHead   = "discovery-decision:v1"
	DecisionClosed          = "CLOSED"
	DecisionUnknown         = "UNKNOWN"
	DecisionRefuted         = "REFUTED"
	CaseExactPairCandidate  = "EXACT_PAIR_CANDIDATE"
	CaseDeterministicReplay = "DETERMINISTIC_REPLAY"
	CaseSingleComparable    = "SINGLE_COMPARABLE_OBSERVATION"
	CaseIncompatibleGroup   = "INCOMPATIBLE_OBSERVATION_GROUP"
	CaseTamperedEntry       = "TAMPERED_LEDGER_ENTRY"
	CaseContradictory       = "CONTRADICTORY_OUTPUTS"
)

var caseIDs = [...]string{
	CaseExactPairCandidate, CaseDeterministicReplay, CaseSingleComparable,
	CaseIncompatibleGroup, CaseTamperedEntry, CaseContradictory,
}

type DecisionCase struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

type Policy struct {
	SourceDigest    string
	EvaluatorDigest string
	ActivityCount   int
	Cases           []DecisionCase
}

func CaseIDs() []string { return append([]string(nil), caseIDs[:]...) }

func GeneratedEvaluatorDigest() string { return generatedpolicy.EvaluatorDigest }

func GeneratedEvaluatorCaseIDs() []string { return generatedpolicy.CaseIDs() }

func PolicySourceDigest() string { return generatedpolicy.PolicySourceDigest }

func Evaluate(caseID string) (string, bool) { return generatedpolicy.Evaluate(caseID) }

func (policy Policy) Decision(caseID string) (string, bool) { return generatedpolicy.Evaluate(caseID) }

func CompileNamed(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("discovery policy source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Policy{}, fmt.Errorf("parse discovery policy: %w", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower discovery policy: %w", err)
	}
	if ir.Package != "selfimprovementdiscovery" || ir.Namespace.String() != "self_improvement_discovery" {
		return Policy{}, fmt.Errorf("discovery policy package/namespace is %q/%q", ir.Package, ir.Namespace)
	}
	var rows []DecisionCase
	activityCount := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != PolicyActivity {
			continue
		}
		activityCount++
		rows, err = parseRows(node.ValueProgram)
		if err != nil {
			return Policy{}, fmt.Errorf("activity %q: %w", PolicyActivity, err)
		}
	}
	if activityCount != 1 {
		return Policy{}, fmt.Errorf("discovery policy activity count = %d, want 1", activityCount)
	}
	return Policy{SourceDigest: cache.HashBytes(source).String(), EvaluatorDigest: evaluatorDigest(rows), ActivityCount: activityCount, Cases: rows}, nil
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
	if policy.EvaluatorDigest != generatedpolicy.EvaluatorDigest || policy.SourceDigest != generatedpolicy.PolicySourceDigest {
		return Policy{}, errors.New("discovery policy and generated evaluator digest differ")
	}
	generatedIDs := generatedpolicy.CaseIDs()
	if len(generatedIDs) != len(policy.Cases) {
		return Policy{}, errors.New("discovery policy and generated evaluator case counts differ")
	}
	for index, row := range policy.Cases {
		if generatedIDs[index] != row.ID {
			return Policy{}, fmt.Errorf("discovery evaluator case %q is not source-bound", row.ID)
		}
		decision, ok := generatedpolicy.Evaluate(row.ID)
		if !ok || decision != row.Decision {
			return Policy{}, fmt.Errorf("discovery evaluator decision for %q is not source-bound", row.ID)
		}
	}
	return policy, nil
}

func parseRows(value string) ([]DecisionCase, error) {
	seenHead := false
	rows := make([]DecisionCase, 0, len(caseIDs))
	seen := make(map[string]bool, len(caseIDs))
	for part := range strings.SplitSeq(value, ";") {
		switch {
		case part == DiscoveryDecisionHead:
			if seenHead {
				return nil, errors.New("discovery decision header is duplicated")
			}
			seenHead = true
		case strings.HasPrefix(part, "discovery-case="):
			encoded := strings.TrimPrefix(part, "discovery-case=")
			id, decision, ok := strings.Cut(encoded, ":")
			if !ok || id == "" || decision == "" || strings.Contains(decision, ":") || seen[id] || !knownCase(id) || !knownDecision(decision) {
				return nil, fmt.Errorf("discovery decision row %q is malformed or outside the fixed policy", encoded)
			}
			seen[id] = true
			rows = append(rows, DecisionCase{ID: id, Decision: decision})
		}
	}
	if !seenHead || len(rows) != len(caseIDs) {
		return nil, fmt.Errorf("discovery decision rows = %d, want %d", len(rows), len(caseIDs))
	}
	for index, id := range caseIDs {
		if rows[index].ID != id {
			return nil, fmt.Errorf("discovery decision row %d is %q, want %q", index+1, rows[index].ID, id)
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
	fmt.Fprintln(&builder, "// Code generated by gooo public discovery policy generator; DO NOT EDIT.")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "package discoverypolicygenerated")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "const (")
	fmt.Fprintf(&builder, "\tEvaluatorSchema = %q\n", EvaluatorSchema)
	fmt.Fprintf(&builder, "\tPolicySourcePath = %q\n", filename)
	fmt.Fprintf(&builder, "\tPolicySourceDigest = %q\n", policy.SourceDigest)
	fmt.Fprintf(&builder, "\tEvaluatorDigest = %q\n", policy.EvaluatorDigest)
	fmt.Fprintln(&builder, ")")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "type DecisionCase struct {")
	fmt.Fprintln(&builder, "\tID string")
	fmt.Fprintln(&builder, "\tDecision string")
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "var cases = [...]DecisionCase{")
	for _, row := range policy.Cases {
		fmt.Fprintf(&builder, "\t{ID: %q, Decision: %q},\n", row.ID, row.Decision)
	}
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "func CaseIDs() []string {")
	fmt.Fprintln(&builder, "\tresult := make([]string, 0, len(cases))")
	fmt.Fprintln(&builder, "\tfor _, row := range cases {")
	fmt.Fprintln(&builder, "\t\tresult = append(result, row.ID)")
	fmt.Fprintln(&builder, "\t}")
	fmt.Fprintln(&builder, "\treturn result")
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "func Evaluate(caseID string) (string, bool) {")
	fmt.Fprintln(&builder, "\tswitch caseID {")
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
		fmt.Fprintf(&builder, "\tcase %q", ids[0])
		for _, id := range ids[1:] {
			fmt.Fprintf(&builder, ", %q", id)
		}
		fmt.Fprintf(&builder, ":\n\t\treturn %q, true\n", decision)
	}
	fmt.Fprintln(&builder, "\tdefault:")
	fmt.Fprintln(&builder, "\t\treturn \"\", false")
	fmt.Fprintln(&builder, "\t}")
	fmt.Fprintln(&builder, "}")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return nil, fmt.Errorf("format generated discovery evaluator: %w", err)
	}
	return formatted, nil
}
