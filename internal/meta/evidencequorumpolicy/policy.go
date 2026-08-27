package evidencequorumpolicy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const policyActivity = "DefineEvidenceQuorumPolicy"

// Policy is lowered from the computes value in policy.gooo. The Go type is a
// schema for values authored in Gooo, not an alternate policy authority.
type Policy struct {
	SemanticDigest         string
	SourcePath             string
	SourceEntry            string
	CaseDenominator        int
	Threshold              int
	RequiredRoles          []string
	RequiredPredicates     []string
	PriorClaimState        string
	ContradictionPredicate string
	Claim                  Claim
	CaseIDs                []string
}

type Claim struct {
	ID            string
	Statement     string
	Producer      string
	Consumer      string
	MetaOperation string
	ProofChoice   string
}

func Parse(sourcePath string, source []byte) (Policy, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return Policy{}, fmt.Errorf("policy source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("policy source lowering failed: %w", err)
	}
	var program string
	declarations := file.Declarations
	if declarations == nil {
		declarations = file.Decls
	}
	for _, declaration := range declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && activity.Name == policyActivity {
			program = activity.ValueProgram
			break
		}
	}
	if program == "" {
		return Policy{}, fmt.Errorf("policy activity %q has no semantic value", policyActivity)
	}
	values := map[string]string{}
	for _, item := range strings.Split(program, ";") {
		key, value, ok := strings.Cut(item, "=")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			return Policy{}, fmt.Errorf("invalid policy value %q", item)
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	threshold, err := intValue(values, "threshold")
	if err != nil {
		return Policy{}, err
	}
	denominator, err := intValue(values, "case_denominator")
	if err != nil {
		return Policy{}, err
	}
	policy := Policy{
		SemanticDigest:         "sha256:" + ir.StableHash(),
		SourcePath:             values["source_path"],
		SourceEntry:            values["source_entry"],
		CaseDenominator:        denominator,
		Threshold:              threshold,
		RequiredRoles:          splitValue(values["required_roles"]),
		RequiredPredicates:     splitValue(values["required_predicates"]),
		PriorClaimState:        values["prior_claim_state"],
		ContradictionPredicate: values["contradiction_predicate"],
		Claim: Claim{
			ID:            values["claim_id"],
			Statement:     values["claim_statement"],
			Producer:      values["producer"],
			Consumer:      values["consumer"],
			MetaOperation: values["meta_operation"],
			ProofChoice:   values["proof_choice"],
		},
		CaseIDs: splitValue(values["cases"]),
	}
	if policy.SourcePath == "" || policy.SourceEntry == "" || policy.Threshold < 1 || policy.CaseDenominator < 1 ||
		policy.PriorClaimState == "" || len(policy.RequiredRoles) == 0 || len(policy.RequiredPredicates) == 0 || len(policy.CaseIDs) != policy.CaseDenominator {
		return Policy{}, fmt.Errorf("policy semantic values are incomplete")
	}
	return policy, nil
}

func intValue(values map[string]string, key string) (int, error) {
	value, err := strconv.Atoi(values[key])
	if err != nil || value < 1 {
		return 0, fmt.Errorf("policy %s must be a positive integer", key)
	}
	return value, nil
}

func splitValue(value string) []string {
	var result []string
	for _, item := range strings.Split(value, "|") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

// SemanticDigest lowers a source independently of source execution and
// intentionally excludes source spans, so comment-only edits preserve it.
func SemanticDigest(filename string, source []byte) (string, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return "", fmt.Errorf("source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", err
	}
	return "sha256:" + ir.StableHash(), nil
}
