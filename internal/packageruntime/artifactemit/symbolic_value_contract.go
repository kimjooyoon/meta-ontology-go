package artifactemit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

const symbolicValueContractSchema = "gooo/symbolic-invocation-value-contract/v1"

type SymbolicValueContract struct {
	Schema               string                           `json:"schema"`
	SubjectSHA           string                           `json:"subject_sha"`
	MetricID             string                           `json:"metric_id"`
	Decision             string                           `json:"decision"`
	Resolution           string                           `json:"resolution"`
	Reason               string                           `json:"reason"`
	SourceArtifactDigest string                           `json:"source_artifact_digest"`
	Rules                []SymbolicValueContractRule      `json:"rules"`
	Default              SymbolicValueContractDefault     `json:"default"`
	Coordinates          SymbolicValueContractCoordinates `json:"coordinates"`
	Classes              []SymbolicValueContractClass     `json:"classes"`
	Indicators           []SymbolicValueContractIndicator `json:"indicators"`
	Views                []SymbolicValueContractView      `json:"views"`
	Proofs               []SymbolicValueContractProof     `json:"proofs"`
	Effects              SymbolicValueContractEffects     `json:"effects"`
	PromotionCreditBPS   int                              `json:"promotion_credit_bps"`
	NotClaimed           []string                         `json:"not_claimed"`
	Digest               string                           `json:"digest,omitempty"`
}

type SymbolicValueContractRule struct {
	ID            string                         `json:"id"`
	Match         SymbolicValueContractRuleMatch `json:"match"`
	Decision      string                         `json:"decision"`
	Resolution    string                         `json:"resolution"`
	Reason        string                         `json:"reason"`
	ProofChoice   string                         `json:"proof_choice"`
	MetaOperation string                         `json:"meta_operation"`
}

type SymbolicValueContractRuleMatch struct {
	Activity string `json:"activity"`
	Inputs   string `json:"inputs"`
}

type SymbolicValueContractDefault struct {
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	Reason        string `json:"reason"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

type SymbolicValueContractCoordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type SymbolicValueContractClass struct {
	Class     string `json:"class"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}

type SymbolicValueContractIndicator struct {
	ID            string   `json:"id"`
	Class         string   `json:"class"`
	ProofChoice   string   `json:"proof_choice"`
	MetaOperation string   `json:"meta_operation"`
	Observed      int      `json:"observed"`
	Expected      int      `json:"expected"`
	Satisfied     bool     `json:"satisfied"`
	Audiences     []string `json:"audiences"`
}

type SymbolicValueContractView struct {
	Audience    string `json:"audience"`
	Resolution  string `json:"resolution"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
	BasisPoints int    `json:"basis_points"`
}

type SymbolicValueContractProof struct {
	ProofChoice string `json:"proof_choice"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
}

type SymbolicValueContractEffects struct {
	RepositoryWrites int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type symbolicValueArtifactInput struct {
	Schema      string `json:"schema"`
	Decision    string `json:"decision"`
	Digest      string `json:"digest"`
	Conformance struct {
		Schema                     string                     `json:"schema"`
		Decision                   string                     `json:"decision"`
		Resolution                 string                     `json:"resolution"`
		GeneratedVectors           int                        `json:"generated_vectors"`
		EmbeddedHandwrittenVectors int                        `json:"embedded_handwritten_vectors"`
		Vectors                    []symbolicValueVectorInput `json:"vectors"`
		Effects                    struct {
			RepositoryWrites int  `json:"repository_writes"`
			MutationAuthority bool `json:"mutation_authority"`
		} `json:"effects"`
	} `json:"conformance"`
}

type symbolicValueVectorInput struct {
	ID            string `json:"id"`
	Expected      string `json:"expected"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Instance      struct {
		Activity string   `json:"activity"`
		Inputs   []string `json:"inputs"`
	} `json:"instance"`
}

func CompileSymbolicValueContract(artifactJSON []byte, subjectSHA string) (SymbolicValueContract, error) {
	var input symbolicValueArtifactInput
	if err := json.Unmarshal(artifactJSON, &input); err != nil {
		return SymbolicValueContract{}, fmt.Errorf("decode symbolic artifact: %w", err)
	}
	if !validSymbolicValueHexDigest(subjectSHA, 20) {
		return SymbolicValueContract{}, fmt.Errorf("subject sha must be 40 lowercase hexadecimal characters")
	}
	if input.Schema != "gooo/symbolic-invocation-schema-artifact/v1" || input.Decision != "PASS" {
		return SymbolicValueContract{}, fmt.Errorf("symbolic artifact identity is not accepted")
	}
	if !validSymbolicValueSHA256(input.Digest) {
		return SymbolicValueContract{}, fmt.Errorf("symbolic artifact digest is not sha256")
	}
	if input.Conformance.Schema != "gooo/symbolic-invocation-conformance/v1" ||
		input.Conformance.Decision != "PASS" ||
		input.Conformance.Resolution != "STRUCTURAL_ONLY" {
		return SymbolicValueContract{}, fmt.Errorf("symbolic conformance identity is not accepted")
	}
	if input.Conformance.GeneratedVectors != 2 || len(input.Conformance.Vectors) != 2 {
		return SymbolicValueContract{}, fmt.Errorf("symbolic conformance must expose exactly two generated vectors")
	}
	if input.Conformance.EmbeddedHandwrittenVectors != 0 {
		return SymbolicValueContract{}, fmt.Errorf("embedded handwritten vectors are not compiler authority")
	}
	if input.Conformance.Effects.RepositoryWrites != 0 || input.Conformance.Effects.MutationAuthority {
		return SymbolicValueContract{}, fmt.Errorf("symbolic conformance must remain read-only")
	}

	var acceptVector *symbolicValueVectorInput
	var rejectVector *symbolicValueVectorInput
	seen := make(map[string]struct{}, len(input.Conformance.Vectors))
	for i := range input.Conformance.Vectors {
		vector := &input.Conformance.Vectors[i]
		if _, duplicate := seen[vector.ID]; duplicate {
			return SymbolicValueContract{}, fmt.Errorf("duplicate symbolic vector %q", vector.ID)
		}
		seen[vector.ID] = struct{}{}
		switch vector.Expected {
		case "ACCEPT", "REJECT":
		default:
			return SymbolicValueContract{}, fmt.Errorf("symbolic vector %q has unknown expected verdict %q", vector.ID, vector.Expected)
		}
		switch vector.ID {
		case "accept-exact":
			acceptVector = vector
		case "reject-missing-activity":
			rejectVector = vector
		default:
			return SymbolicValueContract{}, fmt.Errorf("symbolic vector %q is not contract-bound", vector.ID)
		}
	}
	if acceptVector == nil || rejectVector == nil {
		return SymbolicValueContract{}, fmt.Errorf("required symbolic vectors are missing")
	}
	if acceptVector.Expected != "ACCEPT" || acceptVector.ProofChoice != "FOUNDATION" ||
		acceptVector.MetaOperation != "project-exact-symbolic-invocation" ||
		strings.TrimSpace(acceptVector.Instance.Activity) == "" || len(acceptVector.Instance.Inputs) == 0 {
		return SymbolicValueContract{}, fmt.Errorf("accept vector does not prove a complete symbolic value")
	}
	if rejectVector.Expected != "REJECT" || rejectVector.ProofChoice != "REGRESSION" ||
		rejectVector.MetaOperation != "remove-required-activity" ||
		strings.TrimSpace(rejectVector.Instance.Activity) != "" || len(rejectVector.Instance.Inputs) == 0 {
		return SymbolicValueContract{}, fmt.Errorf("reject vector does not prove the missing-activity boundary")
	}
	if !slices.Equal(acceptVector.Instance.Inputs, rejectVector.Instance.Inputs) {
		return SymbolicValueContract{}, fmt.Errorf("generated vectors do not share an input boundary")
	}

	indicators := []SymbolicValueContractIndicator{
		newSymbolicValueIndicator("compiler.source-artifact-bindings", "DRIVER", "FOUNDATION", "bind-symbolic-schema-artifact-digest", 1, 1, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.generated-vectors", "DRIVER", "FOUNDATION", "count-compiler-generated-contract-vectors", len(input.Conformance.Vectors), 2, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.value-rules", "OUTCOME", "COHERENCE", "compile-symbolic-value-rules", 2, 2, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.rule-mappings", "OUTCOME", "COHERENCE", "map-value-rules-to-decisions", 2, 2, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.default-fail-closed-policies", "DRIVER", "REGRESSION", "compile-unmatched-value-fail-closed-default", 1, 1, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("compiler.embedded-handwritten-vectors", "GUARDRAIL", "REGRESSION", "count-embedded-handwritten-contract-vectors", input.Conformance.EmbeddedHandwrittenVectors, 0, "TOOL_AUTHOR", "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.repository-writes", "GUARDRAIL", "FOUNDATION", "sum-value-contract-repository-writes", 0, 0, "GOVERNOR"),
		newSymbolicValueIndicator("guardrail.mutation-authorities", "GUARDRAIL", "FOUNDATION", "join-value-contract-mutation-authority", 0, 0, "GOVERNOR"),
	}

	contract := SymbolicValueContract{
		Schema:               symbolicValueContractSchema,
		SubjectSHA:           subjectSHA,
		MetricID:             "gooo.metric.compiler.symbolic-value-contract.v1",
		Decision:             "PASS",
		Resolution:           "VALUE_CONTRACT_ONLY",
		Reason:               "SYMBOLIC_VALUE_SEMANTICS_COMPILED",
		SourceArtifactDigest: input.Digest,
		Rules: []SymbolicValueContractRule{
			{
				ID:            "complete-symbolic-invocation",
				Match:         SymbolicValueContractRuleMatch{Activity: "NON_EMPTY", Inputs: "NON_EMPTY"},
				Decision:      "READY",
				Resolution:    "VALUE_EXACT",
				Reason:        "SYMBOLIC_INVOCATION_VALUE_PROJECTED",
				ProofChoice:   acceptVector.ProofChoice,
				MetaOperation: acceptVector.MetaOperation,
			},
			{
				ID:            "missing-activity",
				Match:         SymbolicValueContractRuleMatch{Activity: "MISSING_OR_EMPTY", Inputs: "ANY"},
				Decision:      "FAIL_CLOSED",
				Resolution:    "LOWER_RESOLUTION",
				Reason:        "SYMBOLIC_INVOCATION_VALUE_INCOMPLETE",
				ProofChoice:   rejectVector.ProofChoice,
				MetaOperation: rejectVector.MetaOperation,
			},
		},
		Default: SymbolicValueContractDefault{
			Decision:      "FAIL_CLOSED",
			Resolution:    "LOWER_RESOLUTION",
			Reason:        "SYMBOLIC_INVOCATION_VALUE_UNMATCHED",
			ProofChoice:   "REGRESSION",
			MetaOperation: "fail-closed-unmatched-symbolic-value",
		},
		Indicators: indicators,
		Effects: SymbolicValueContractEffects{
			RepositoryWrites: 0,
			MutationAuthority: false,
		},
		PromotionCreditBPS: 0,
		NotClaimed: []string{
			"default-policy external coverage",
			"effect execution",
			"arbitrary user input",
			"complete interpreter semantics",
			"domain correctness",
			"production readiness",
		},
	}
	contract.Coordinates = symbolicValueCoordinates(indicators)
	contract.Classes = symbolicValueClasses(indicators)
	contract.Views = []SymbolicValueContractView{
		symbolicValueView(indicators, "USER", "USER_VISIBLE"),
		symbolicValueView(indicators, "TOOL_AUTHOR", "TOOL_CONTRACT"),
		symbolicValueView(indicators, "GOVERNOR", "FULL_RECEIPT"),
	}
	contract.Proofs = symbolicValueProofs(indicators)
	if contract.Coordinates.Satisfied != contract.Coordinates.Total {
		contract.Decision = "FAIL_CLOSED"
		contract.Resolution = "INVARIANT_ONLY"
		contract.Reason = "SYMBOLIC_VALUE_CONTRACT_INCOMPLETE"
	}
	canonical, err := canonicalSymbolicValueContract(contract)
	if err != nil {
		return SymbolicValueContract{}, fmt.Errorf("canonicalize symbolic value contract: %w", err)
	}
	digest := sha256.Sum256(canonical)
	contract.Digest = "sha256:" + hex.EncodeToString(digest[:])
	return contract, nil
}

func newSymbolicValueIndicator(id, class, proof, operation string, observed, expected int, audiences ...string) SymbolicValueContractIndicator {
	return SymbolicValueContractIndicator{
		ID:            id,
		Class:         class,
		ProofChoice:   proof,
		MetaOperation: operation,
		Observed:      observed,
		Expected:      expected,
		Satisfied:     observed == expected,
		Audiences:     audiences,
	}
}

func symbolicValueCoordinates(indicators []SymbolicValueContractIndicator) SymbolicValueContractCoordinates {
	satisfied := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	total := len(indicators)
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10000 / total
	}
	return SymbolicValueContractCoordinates{Satisfied: satisfied, Total: total, BasisPoints: basisPoints}
}

func symbolicValueClasses(indicators []SymbolicValueContractIndicator) []SymbolicValueContractClass {
	classes := []string{"OUTCOME", "DRIVER", "GUARDRAIL"}
	result := make([]SymbolicValueContractClass, 0, len(classes))
	for _, class := range classes {
		total := 0
		satisfied := 0
		for _, indicator := range indicators {
			if indicator.Class != class {
				continue
			}
			total++
			if indicator.Satisfied {
				satisfied++
			}
		}
		result = append(result, SymbolicValueContractClass{Class: class, Satisfied: satisfied, Total: total})
	}
	return result
}

func symbolicValueView(indicators []SymbolicValueContractIndicator, audience, resolution string) SymbolicValueContractView {
	total := 0
	satisfied := 0
	for _, indicator := range indicators {
		if !slices.Contains(indicator.Audiences, audience) {
			continue
		}
		total++
		if indicator.Satisfied {
			satisfied++
		}
	}
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10000 / total
	}
	return SymbolicValueContractView{
		Audience: audience, Resolution: resolution, Satisfied: satisfied, Total: total, BasisPoints: basisPoints,
	}
}

func symbolicValueProofs(indicators []SymbolicValueContractIndicator) []SymbolicValueContractProof {
	proofs := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	result := make([]SymbolicValueContractProof, 0, len(proofs))
	for _, proof := range proofs {
		total := 0
		satisfied := 0
		for _, indicator := range indicators {
			if indicator.ProofChoice != proof {
				continue
			}
			total++
			if indicator.Satisfied {
				satisfied++
			}
		}
		result = append(result, SymbolicValueContractProof{ProofChoice: proof, Satisfied: satisfied, Total: total})
	}
	return result
}

func canonicalSymbolicValueContract(contract SymbolicValueContract) ([]byte, error) {
	contract.Digest = ""
	raw, err := json.Marshal(contract)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

func validSymbolicValueSHA256(value string) bool {
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	return validSymbolicValueHexDigest(strings.TrimPrefix(value, "sha256:"), 32)
}

func validSymbolicValueHexDigest(value string, bytesLength int) bool {
	if value != strings.ToLower(value) || len(value) != bytesLength*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == bytesLength
}
