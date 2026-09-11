package policycompilation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// DeclaredCaseEvaluation separates conditional Gooo policy evaluation from
// verification of the caller's evidence claims. It is not a conformance receipt.
type DeclaredCaseEvaluation struct {
	Schema                     string              `json:"schema"`
	SourceDigest               string              `json:"source_digest"`
	SemanticDigest             string              `json:"semantic_digest"`
	InputDigest                string              `json:"input_digest"`
	EffectiveInputDigest       string              `json:"effective_input_digest"`
	EvaluationScope            string              `json:"evaluation_scope"`
	ExternalEvidenceState      string              `json:"external_evidence_state"`
	FullConformanceState       string              `json:"full_conformance_state"`
	GeneratedExecutionObserved bool                `json:"generated_execution_observed"`
	MutationAuthority          int                 `json:"mutation_authority"`
	PromotionAuthority         int                 `json:"promotion_authority"`
	SourceDecision             DecisionResult      `json:"source_decision"`
	Fields                     []DeclaredCaseField `json:"fields"`
}

type DeclaredCaseField struct {
	JSONField string          `json:"json_field"`
	GoType    string          `json:"go_type"`
	Binding   string          `json:"binding"`
	Supplied  json.RawMessage `json:"supplied,omitempty"`
	Effective json.RawMessage `json:"effective"`
}

// EvaluateDeclaredCase compiles the supplied Gooo source and preserves how each
// Case field acquired its effective value. It neither executes generated Go nor
// verifies caller-supplied availability, provenance, or evidence digests.
func EvaluateDeclaredCase(filename string, source, document []byte, expectedPackage, expectedNamespace string) (DeclaredCaseEvaluation, error) {
	policy, err := CompileForIdentity(filename, source, expectedPackage, expectedNamespace)
	if err != nil {
		return DeclaredCaseEvaluation{}, fmt.Errorf("compile declared-case policy: %w", err)
	}
	var input Case
	if err := decodeStrictJSON(document, &input); err != nil {
		return DeclaredCaseEvaluation{}, fmt.Errorf("decode declared case: %w", err)
	}
	var supplied map[string]json.RawMessage
	if err := json.Unmarshal(document, &supplied); err != nil || supplied == nil {
		return DeclaredCaseEvaluation{}, errors.New("declared case must be a JSON object")
	}
	fields, err := bindDeclaredCaseInputFields(input, supplied)
	if err != nil {
		return DeclaredCaseEvaluation{}, err
	}
	effective, err := json.Marshal(input)
	if err != nil {
		return DeclaredCaseEvaluation{}, fmt.Errorf("encode effective declared case: %w", err)
	}
	return DeclaredCaseEvaluation{
		Schema:                     DeclaredCaseEvaluationSchema,
		SourceDigest:               policy.SourceDigest,
		SemanticDigest:             policy.SemanticDigest,
		InputDigest:                DigestBytes(document),
		EffectiveInputDigest:       DigestBytes(effective),
		EvaluationScope:            "CONDITIONAL_SOURCE_POLICY_EVALUATION",
		ExternalEvidenceState:      "UNKNOWN_NOT_VERIFIED",
		FullConformanceState:       "UNKNOWN_NOT_EXECUTED",
		GeneratedExecutionObserved: false,
		MutationAuthority:          0,
		PromotionAuthority:         0,
		SourceDecision:             EvaluateSourcePolicy(policy, input),
		Fields:                     fields,
	}, nil
}

func bindDeclaredCaseInputFields(input Case, supplied map[string]json.RawMessage) ([]DeclaredCaseField, error) {
	schema, value := reflect.TypeFor[Case](), reflect.ValueOf(input)
	fields := make([]DeclaredCaseField, 0, schema.NumField())
	known := make(map[string]bool, schema.NumField())
	for index := 0; index < schema.NumField(); index++ {
		field := schema.Field(index)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if field.PkgPath != "" || name == "-" {
			continue
		}
		if field.Anonymous {
			return nil, errors.New("embedded Case fields require an explicit declaration binding")
		}
		if name == "" {
			name = field.Name
		}
		known[name] = true
		effective, err := json.Marshal(value.Field(index).Interface())
		if err != nil {
			return nil, fmt.Errorf("encode effective Case field %q: %w", name, err)
		}
		raw, present := supplied[name]
		binding := "DECLARED"
		if !present {
			binding = "MISSING_DEFAULTED"
		} else if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			binding = "NULL_DEFAULTED"
		}
		fields = append(fields, DeclaredCaseField{
			JSONField: name,
			GoType:    field.Type.String(),
			Binding:   binding,
			Supplied:  append(json.RawMessage(nil), raw...),
			Effective: effective,
		})
	}
	// encoding/json accepts case-folded aliases. This explanatory API requires
	// exact field names so a supplied alias cannot be mislabeled as missing.
	unknown := []string{}
	for name := range supplied {
		if !known[name] {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		return nil, fmt.Errorf("declared Case field %q has no exact JSON binding", unknown[0])
	}
	return fields, nil
}

const (
	DeclaredCaseEvaluationSchema       = "gooo/meta-policy-declared-case-evaluation/v1"
	PublicProfileID                    = "meta-policy-compilation-v3"
	PublicGenerationManifestSchema     = "gooo/meta-policy-compilation-generation-manifest/v1"
	PublicGenerationOutputRootClass    = "CALLER_OWNED_EXTERNAL"
	PublicGenerationConformanceUnknown = "UNKNOWN"
)

// PublicGenerationManifest binds the source-derived policy and the generated
// standalone judge without claiming that either has been executed. Execution
// and conformance are deliberately CI-owned concerns.
type PublicGenerationManifest struct {
	Schema               string   `json:"schema"`
	Profile              string   `json:"profile"`
	SourceFile           string   `json:"source_file"`
	SourceDigest         string   `json:"source_digest"`
	SemanticDigest       string   `json:"semantic_digest"`
	Package              string   `json:"package"`
	Namespace            string   `json:"namespace"`
	PolicyBytesDigest    string   `json:"policy_bytes_digest"`
	ArtifactBytesDigest  string   `json:"artifact_bytes_digest"`
	GeneratedJudgeDigest string   `json:"generated_judge_digest"`
	GeneratedFiles       []string `json:"generated_files"`
	OutputRootClass      string   `json:"output_root_class"`
	ExecutionObserved    bool     `json:"execution_observed"`
	CurrentConformance   string   `json:"current_conformance"`
	RepositoryWrites     int      `json:"repository_writes"`
	MutationAuthority    int      `json:"mutation_authority"`
	PromotionAuthority   int      `json:"promotion_authority"`
}
