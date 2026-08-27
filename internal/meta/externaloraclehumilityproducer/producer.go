// Package externaloraclehumilityproducer emits a source-bound wire receipt.
// It deliberately does not import the consumer or any contract model.
package externaloraclehumilityproducer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const ReceiptSchema = "gooo/external-oracle-humility-receipt/v2"

type Claim struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type SourcePolicy struct {
	SourceAuthority           string  `json:"source_authority"`
	ExternalEvidenceRelation  string  `json:"external_evidence_relation"`
	ExternalEvidenceAuthority string  `json:"external_evidence_authority"`
	Claims                    []Claim `json:"claims"`
}

type Declaration struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	ValueProgram string `json:"value_program,omitempty"`
}

type SourceReceipt struct {
	Schema         string        `json:"schema"`
	SubjectSHA     string        `json:"subject_sha"`
	SourcePath     string        `json:"source_path"`
	SourceSHA256   string        `json:"source_sha256"`
	SemanticSHA256 string        `json:"semantic_sha256"`
	Producer       string        `json:"producer"`
	Consumer       string        `json:"consumer"`
	MetaOperation  string        `json:"meta_operation"`
	ProofChoice    string        `json:"proof_choice"`
	Stage          string        `json:"stage"`
	Step           string        `json:"step"`
	Reason         string        `json:"reason"`
	LowerPipeline  []string      `json:"lower_pipeline"`
	Declarations   []Declaration `json:"declarations"`
	SourcePolicy   SourcePolicy  `json:"source_policy"`
}

func ProduceSourceReceipt(subject, sourcePath string, source []byte) (SourceReceipt, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() {
		return SourceReceipt{}, fmt.Errorf("producer parse %s: syntax diagnostics", sourcePath)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceReceipt{}, fmt.Errorf("producer lower %s: %w", sourcePath, err)
	}
	declarations := make([]Declaration, 0, len(ir.Graph.Nodes()))
	for _, node := range ir.Graph.Nodes() {
		declarations = append(declarations, Declaration{
			ID: node.ID.String(), Kind: node.Kind.String(), Name: node.Name,
			ValueProgram: node.ValueProgram,
		})
	}
	policy, err := policyFromDeclarations(declarations)
	if err != nil {
		return SourceReceipt{}, fmt.Errorf("producer extract source policy: %w", err)
	}
	return SourceReceipt{
		Schema: ReceiptSchema, SubjectSHA: subject, SourcePath: sourcePath,
		SourceSHA256: DigestBytes(source), SemanticSHA256: DigestString(ir.SemanticCanonical()),
		Producer: "source-receipt-producer", Consumer: "external-oracle-humility-consumer",
		MetaOperation: "emit-source-receipt", ProofChoice: "FOUNDATION",
		Stage: "observe", Step: "source-receipt", Reason: "SOURCE_POLICY_BOUND",
		LowerPipeline: []string{"syntax.ParseFile", "bidir.Lower"},
		Declarations:  declarations, SourcePolicy: policy,
	}, nil
}

func policyFromDeclarations(declarations []Declaration) (SourcePolicy, error) {
	var policy SourcePolicy
	for _, declaration := range declarations {
		if declaration.Kind != "Activity" || declaration.ValueProgram == "" {
			continue
		}
		values, err := parseProgram(declaration.ValueProgram)
		if err != nil {
			return SourcePolicy{}, fmt.Errorf("activity %s: %w", declaration.Name, err)
		}
		switch declaration.Name {
		case "ComputeSourceAuthorityPolicy":
			policy.SourceAuthority = values["source_authority"]
			policy.ExternalEvidenceRelation = values["external_evidence_relation"]
			policy.ExternalEvidenceAuthority = values["external_evidence_authority"]
		default:
			claimID, okID := values["claim_id"]
			state, okState := values["state"]
			if okID && okState {
				policy.Claims = append(policy.Claims, Claim{ID: claimID, State: state})
			}
		}
	}
	if policy.SourceAuthority == "" || policy.ExternalEvidenceRelation == "" || policy.ExternalEvidenceAuthority == "" {
		return SourcePolicy{}, fmt.Errorf("source authority policy is absent from computes values")
	}
	sort.Slice(policy.Claims, func(i, j int) bool { return policy.Claims[i].ID < policy.Claims[j].ID })
	return policy, nil
}

func parseProgram(program string) (map[string]string, error) {
	values := make(map[string]string)
	for _, item := range strings.Split(program, ";") {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("invalid computes field %q", item)
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("duplicate computes field %q", key)
		}
		values[key] = value
	}
	return values, nil
}

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestString(value string) string { return DigestBytes([]byte(value)) }
