package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const sourceObservationSchema = "gooo/meta-policy-source-observation/v1"

type sourceObservationRequest struct {
	PolicyPath        string
	ExpectedPackage   string
	ExpectedNamespace string
	Flags             []string
	Arguments         []string
}

type sourceObservationUnknown struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type policySourceObservation struct {
	Schema                    string                   `json:"schema"`
	PolicyID                  string                   `json:"policy_id"`
	Package                   string                   `json:"package"`
	Namespace                 string                   `json:"namespace"`
	SourceDigest              string                   `json:"source_digest"`
	SemanticDigest            string                   `json:"semantic_digest"`
	Denominator               int                      `json:"fixed_denominator"`
	RuleBindings              []rule                   `json:"rule_bindings"`
	DecisionRules             []decisionRule           `json:"decision_rules"`
	SourceParseObserved       bool                     `json:"source_parse_observed"`
	LoweringObserved          bool                     `json:"lowering_observed"`
	PolicyExecutionObserved   bool                     `json:"policy_execution_observed"`
	ProducerArtifactsObserved bool                     `json:"producer_artifacts_observed"`
	CurrentConformance        string                   `json:"current_conformance"`
	ConformanceUnknown        sourceObservationUnknown `json:"conformance_unknown"`
	RepositoryWrites          int                      `json:"repository_writes"`
	MutationAuthority         int                      `json:"mutation_authority"`
	PromotionAuthority        int                      `json:"promotion_authority"`
}

func validateSourceObservationRequest(request sourceObservationRequest) error {
	if request.PolicyPath == "" {
		return errors.New("source observation requires -policy")
	}
	if request.ExpectedPackage == "" || request.ExpectedNamespace == "" {
		return errors.New("source observation requires an expected package and namespace")
	}
	if len(request.Arguments) != 0 {
		return errors.New("source observation does not accept positional arguments")
	}
	for _, name := range request.Flags {
		switch name {
		case "observe-source", "policy", "profile-package", "profile-namespace":
		default:
			return fmt.Errorf("source observation does not accept flag -%s", name)
		}
	}
	return nil
}

func reconstructSourceObservation(filename string, source []byte, expectedPackage, expectedNamespace string) (policySourceObservation, error) {
	compiled, err := parseRawPolicy(filename, source, expectedPackage, expectedNamespace)
	if err != nil {
		return policySourceObservation{}, fmt.Errorf("reconstruct raw policy source: %w", err)
	}
	return policySourceObservation{
		Schema: sourceObservationSchema, PolicyID: compiled.PolicyID,
		Package: compiled.Package, Namespace: compiled.Namespace,
		SourceDigest: compiled.SourceDigest, SemanticDigest: compiled.SemanticDigest,
		Denominator: compiled.Denominator, RuleBindings: compiled.Rules, DecisionRules: compiled.Reduction.Rules,
		SourceParseObserved: true, LoweringObserved: true, CurrentConformance: decisionUnknown,
		ConformanceUnknown: sourceObservationUnknown{
			State: decisionUnknown, Stage: "POLICY_CONFORMANCE", Step: "OBSERVE_RAW_SOURCE",
			Reason: "GENERATED_EXECUTION_NOT_OBSERVED", UnknownClass: "DIRECT_MISSING",
			NextOperation: "OBSERVE_SOURCE_BOUND_GENERATED_EXECUTION", BlockedBy: []string{},
		},
	}, nil
}

func runSourceObservation(request sourceObservationRequest, output io.Writer) error {
	if err := validateSourceObservationRequest(request); err != nil {
		return err
	}
	source, err := os.ReadFile(request.PolicyPath)
	if err != nil {
		return fmt.Errorf("read raw policy source: %w", err)
	}
	observed, err := reconstructSourceObservation(request.PolicyPath, source, request.ExpectedPackage, request.ExpectedNamespace)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(observed); err != nil {
		return fmt.Errorf("emit source observation: %w", err)
	}
	return nil
}
