package compilercompatibility

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
)

func EvaluateStrict(predecessor, successor ExecutionReceipt) Evaluation {
	comparisons := compareAxes(predecessor.IdentityAxes(), successor.IdentityAxes())
	if allEqual(comparisons) {
		return Evaluation{Decision: DecisionClosed, Reason: ReasonExactStrictReplay, Axes: comparisons, ExactSubjectBinding: true, CompatibilityHit: true}
	}
	return Evaluation{Decision: DecisionRefuted, Reason: ReasonAxisMismatch, Axes: comparisons, MismatchDetected: true}
}

func EvaluateOptIn(policy compatibilitypolicy.Policy, request Request) Evaluation {
	if request.Certificate == nil {
		return unknown(ReasonMissingCertificate, "LOAD_COMPATIBILITY_CERTIFICATE", "compatibility_certificate", "PROVIDE_BOUNDED_SUCCESSOR_CERTIFICATE")
	}
	certificate := request.Certificate
	if request.Current.CandidateStableID != request.CandidateStableID || request.Current.SubjectDigest != request.SubjectDigest || request.Current.SourceDigest != request.SourceDigest {
		return Evaluation{Decision: DecisionRefuted, Reason: ReasonCurrentSubjectMismatch, FallbackAttempted: true, FallbackRejected: true, MismatchDetected: true}
	}
	if certificate.CandidateStableID != request.CandidateStableID || certificate.SubjectDigest != request.SubjectDigest || certificate.SourceDigest != request.SourceDigest {
		return Evaluation{Decision: DecisionRefuted, Reason: ReasonCurrentSubjectMismatch, FallbackAttempted: true, FallbackRejected: true, MismatchDetected: true}
	}
	if certificate.ScopeBounded && len(certificate.Scope) > 1 {
		return Evaluation{Decision: DecisionRefuted, Reason: ReasonWidenedScope, MismatchDetected: true}
	}
	if !certificate.ScopeBounded {
		return unknown(ReasonUnboundedScope, "CHECK_COMPATIBILITY_SCOPE", "unbounded_compatibility_scope", "NARROW_SCOPE_TO_ONE_CANONICAL_SUBJECT")
	}
	if certificate.Successor.Role == "" || certificate.Successor.SemanticIRDigest == "" {
		return unknown(ReasonMissingSuccessorReplay, "VERIFY_SUCCESSOR_REPLAY", "successor_replay_evidence", "PROVIDE_INDEPENDENT_SUCCESSOR_REPLAY")
	}
	if err := ValidateCertificate(*certificate); err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) && (validationErr.Kind == ValidationMissingReplay || validationErr.Kind == ValidationMissingSuccessor || validationErr.Kind == ValidationUnboundedScope || validationErr.Kind == ValidationAmbiguousAxes) {
			return unknown(ReasonAmbiguousEvidence, "VERIFY_SUCCESSOR_REPLAY", "independent_replay_evidence", "RECORD_COMPLETE_INDEPENDENT_REPLAY")
		}
		return Evaluation{Decision: DecisionRefuted, Reason: reasonForValidation(err), MismatchDetected: true}
	}
	if certificate.PolicyDigest != policy.SourceDigest || certificate.PolicyEvaluatorDigest != policy.EvaluatorDigest {
		return Evaluation{Decision: DecisionRefuted, Reason: ReasonAxisMismatch, MismatchDetected: true}
	}
	currentAxes := request.Current.IdentityAxes()
	successorAxes := certificate.SuccessorAxes
	comparisons := compareAxes(successorAxes, currentAxes)
	if !allEqual(comparisons) {
		return Evaluation{Decision: DecisionRefuted, Reason: ReasonAxisMismatch, Axes: comparisons, MismatchDetected: true}
	}
	if request.Current.SemanticIRDigest != certificate.Successor.SemanticIRDigest ||
		request.Current.GeneratedOutputDigest != certificate.Successor.GeneratedOutputDigest ||
		request.Current.GeneratedManifestDigest != certificate.Successor.GeneratedManifestDigest ||
		!reflect.DeepEqual(request.Current.GeneratedSource, certificate.Successor.GeneratedSource) ||
		!reflect.DeepEqual(request.Current.GeneratedManifest, certificate.Successor.GeneratedManifest) {
		return Evaluation{Decision: DecisionRefuted, Reason: ReasonAxisMismatch, Axes: comparisons, MismatchDetected: true}
	}
	return Evaluation{Decision: DecisionClosed, Reason: ReasonBoundedSuccessorReplay, Axes: comparisons, ExactSubjectBinding: true,
		IndependentReplayExecutions: certificate.IndependentReplayExecutions, CompatibilityHit: true}
}

func compareAxes(predecessor, successor IdentityAxes) []AxisComparison {
	values := []struct {
		name  string
		left  string
		right string
	}{
		{"SEMANTIC_IDENTITY", predecessor.SemanticIdentity, successor.SemanticIdentity},
		{"COMPILER_IMPLEMENTATION_IDENTITY", predecessor.CompilerImplementation, successor.CompilerImplementation},
		{"GO_TOOLCHAIN_IDENTITY", predecessor.GoToolchain, successor.GoToolchain},
		{"POLICY_IDENTITY", predecessor.PolicyIdentity, successor.PolicyIdentity},
		{"GENERATED_ARTIFACT_IDENTITY", predecessor.GeneratedArtifactIdentity, successor.GeneratedArtifactIdentity},
		{"TEST_CONTRACT_IDENTITY", predecessor.TestContractIdentity, successor.TestContractIdentity},
		{"AUTHORIZATION_IDENTITY", predecessor.AuthorizationIdentity, successor.AuthorizationIdentity},
	}
	result := make([]AxisComparison, 0, len(values))
	for _, value := range values {
		result = append(result, AxisComparison{Axis: value.name, Predecessor: value.left, Successor: value.right, Equal: value.left == value.right})
	}
	return result
}

func CompareAxes(predecessor, successor IdentityAxes) []AxisComparison {
	return compareAxes(predecessor, successor)
}

func allEqual(comparisons []AxisComparison) bool {
	for _, comparison := range comparisons {
		if !comparison.Equal {
			return false
		}
	}
	return true
}

func unknown(reason, step, blockedBy, next string) Evaluation {
	return Evaluation{Decision: DecisionUnknown, Reason: reason, Unknown: &UnknownState{Stage: "COMPATIBILITY", Step: step, Reason: reason,
		UnknownClass: "INCOMPLETE_EVIDENCE", NextOperation: next, BlockedBy: []string{blockedBy}}}
}

func reasonForValidation(err error) string {
	if validationErr, ok := errors.AsType[*ValidationError](err); ok {
		switch validationErr.Kind {
		case ValidationTampered:
			return ReasonTamperedCertificate
		case ValidationWidenedScope:
			return ReasonWidenedScope
		case ValidationAxisMismatch:
			return ReasonAxisMismatch
		}
	}
	return fmt.Sprintf("%s: %v", ReasonAxisMismatch, err)
}

func FixedPolicyCaseDecision(policy compatibilitypolicy.Policy, caseID string) (string, error) {
	decision, ok := policy.Decision(caseID)
	if !ok {
		return "", fmt.Errorf("policy has no compatibility case %q", caseID)
	}
	return decision, nil
}
