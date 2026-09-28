package compilercompatibility

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
)

const (
	ValidationTampered         = "TAMPERED"
	ValidationWidenedScope     = "WIDENED_SCOPE"
	ValidationUnboundedScope   = "UNBOUNDED_SCOPE"
	ValidationMissingReplay    = "MISSING_REPLAY"
	ValidationMissingSuccessor = "MISSING_SUCCESSOR_EVIDENCE"
	ValidationAmbiguousAxes    = "AMBIGUOUS_AXES"
	ValidationAxisMismatch     = "AXIS_MISMATCH"
)

type ValidationError struct {
	Kind string
	Err  error
}

func (err *ValidationError) Error() string { return err.Kind + ": " + err.Err.Error() }

func (err *ValidationError) Unwrap() error { return err.Err }

func validation(kind, message string) error {
	return &ValidationError{Kind: kind, Err: errors.New(message)}
}

func ValidateExecutionReceipt(receipt ExecutionReceipt) error {
	if receipt.Schema != ConsumptionSchema || receipt.Role == "" || receipt.CandidateStableID == "" ||
		!validDigest(receipt.SubjectDigest) || !validDigest(receipt.SourceDigest) ||
		receipt.SourceDigest != receipt.SubjectDigest ||
		!validDigest(receipt.SemanticIRDigest) || !validDigest(receipt.GeneratedOutputDigest) ||
		!validDigest(receipt.GeneratedManifestDigest) || len(receipt.GeneratedSource) == 0 || len(receipt.GeneratedManifest) == 0 ||
		!validDigest(receipt.PolicyDigest) || !validDigest(receipt.PolicyEvaluatorDigest) || receipt.PolicyResult != DecisionClosed ||
		!validDigest(receipt.CompilerImplementationDigest) || !validDigest(receipt.GoToolchainDigest) ||
		!validDigest(receipt.TestContractDigest) || !validDigest(receipt.AuthorizationDigest) {
		return validation(ValidationMissingSuccessor, "execution receipt is incomplete")
	}
	if receipt.TestContractResult == "" {
		return validation(ValidationMissingReplay, "execution receipt has no test-contract result")
	}
	if cache.HashBytes(receipt.GeneratedSource).String() != receipt.GeneratedOutputDigest ||
		cache.HashBytes(receipt.GeneratedManifest).String() != receipt.GeneratedManifestDigest {
		return validation(ValidationTampered, "execution receipt bytes do not match their digests")
	}
	axes := receipt.IdentityAxes()
	if !axes.Complete() {
		return validation(ValidationAmbiguousAxes, "execution receipt identity axes are incomplete")
	}
	return nil
}

func ValidateAuthorization(authorization Authorization) error {
	if authorization.Schema != AuthorizationSchema || authorization.AuthorizationID == "" || !validDigest(authorization.AuthorizationID) ||
		authorization.Mode != AuthorizationMode || authorization.CandidateStableID == "" || !validDigest(authorization.SubjectDigest) ||
		!validDigest(authorization.SuccessorCompilerDigest) || !authorization.Authorized || authorization.TransitiveCompatibility ||
		!authorization.ScopeBounded || len(authorization.Scope) != 1 || authorization.Scope[0].CandidateStableID != authorization.CandidateStableID ||
		authorization.Scope[0].SubjectDigest != authorization.SubjectDigest {
		return validation(ValidationTampered, "authorization is not a bounded immutable successor scope")
	}
	digest, err := authorization.ContentDigest()
	if err != nil || digest != authorization.AuthorizationID {
		return validation(ValidationTampered, "authorization is not content-addressed")
	}
	return nil
}

func ValidateCertificate(certificate Certificate) error {
	if certificate.Schema != CertificateSchema || certificate.Mode != CertificateMode || certificate.CertificateID == "" || !validDigest(certificate.CertificateID) ||
		certificate.CandidateStableID == "" || !validDigest(certificate.SubjectDigest) || !validDigest(certificate.SourceDigest) ||
		!validDigest(certificate.PolicyDigest) || !validDigest(certificate.PolicyEvaluatorDigest) ||
		certificate.RepositoryWrites != 0 || certificate.LocalTestExecutions != 0 || certificate.IndependentReplayExecutions != 2 ||
		!certificate.ScopeBounded || certificate.Successor.Role == "" || certificate.Predecessor.Role == "" {
		return validation(ValidationMissingSuccessor, "compatibility certificate envelope is incomplete")
	}
	digest, err := certificate.ContentDigest()
	if err != nil || digest != certificate.CertificateID {
		return validation(ValidationTampered, "compatibility certificate is not content-addressed")
	}
	if len(certificate.Scope) == 0 {
		return validation(ValidationUnboundedScope, "compatibility certificate has no bounded scope")
	}
	if len(certificate.Scope) != 1 {
		return validation(ValidationWidenedScope, "compatibility certificate scope contains more than one subject")
	}
	if certificate.TransitiveCompatibility() {
		return validation(ValidationWidenedScope, "transitive compatibility is not supported")
	}
	if err := ValidateAuthorization(certificate.Authorization); err != nil {
		return err
	}
	if certificate.AuthorizationDigest != certificate.Authorization.AuthorizationID {
		return validation(ValidationTampered, "certificate authorization digest differs from authorization identity")
	}
	if err := ValidateExecutionReceipt(certificate.Predecessor); err != nil {
		return err
	}
	if err := ValidateExecutionReceipt(certificate.Successor); err != nil {
		return err
	}
	if certificate.Predecessor.Role == certificate.Successor.Role ||
		certificate.CandidateStableID != certificate.Predecessor.CandidateStableID || certificate.CandidateStableID != certificate.Successor.CandidateStableID ||
		certificate.SubjectDigest != certificate.Predecessor.SubjectDigest || certificate.SubjectDigest != certificate.Successor.SubjectDigest ||
		certificate.SourceDigest != certificate.Predecessor.SourceDigest || certificate.SourceDigest != certificate.Successor.SourceDigest ||
		certificate.PolicyDigest != certificate.Predecessor.PolicyDigest || certificate.PolicyDigest != certificate.Successor.PolicyDigest ||
		certificate.PolicyEvaluatorDigest != certificate.Predecessor.PolicyEvaluatorDigest || certificate.PolicyEvaluatorDigest != certificate.Successor.PolicyEvaluatorDigest {
		return validation(ValidationAxisMismatch, "certificate execution identities do not share the canonical subject and policy")
	}
	if certificate.PredecessorReceiptDigest != receiptDigest(certificate.Predecessor) || certificate.SuccessorReceiptDigest != receiptDigest(certificate.Successor) {
		return validation(ValidationTampered, "certificate receipt digests do not bind the embedded executions")
	}
	if certificate.Authorization.CandidateStableID != certificate.CandidateStableID || certificate.Authorization.SubjectDigest != certificate.SubjectDigest ||
		certificate.Authorization.SuccessorCompilerDigest != certificate.Successor.CompilerImplementationDigest {
		return validation(ValidationTampered, "authorization does not bind the exact successor")
	}
	if !reflect.DeepEqual(certificate.Scope, certificate.Authorization.Scope) || certificate.Scope[0].CandidateStableID != certificate.CandidateStableID || certificate.Scope[0].SubjectDigest != certificate.SubjectDigest {
		return validation(ValidationTampered, "certificate scope is not authorization-bound")
	}
	if certificate.Predecessor.TestContractResult != "PASS" || certificate.Successor.TestContractResult != "PASS" ||
		certificate.Predecessor.TestContractDigest != certificate.Successor.TestContractDigest ||
		certificate.Predecessor.GoToolchainDigest != certificate.Successor.GoToolchainDigest ||
		certificate.Predecessor.SemanticIRDigest != certificate.Successor.SemanticIRDigest ||
		certificate.Predecessor.GeneratedOutputDigest != certificate.Successor.GeneratedOutputDigest ||
		certificate.Predecessor.GeneratedManifestDigest != certificate.Successor.GeneratedManifestDigest ||
		!reflect.DeepEqual(certificate.Predecessor.GeneratedSource, certificate.Successor.GeneratedSource) ||
		!reflect.DeepEqual(certificate.Predecessor.GeneratedManifest, certificate.Successor.GeneratedManifest) ||
		certificate.Predecessor.PolicyResult != certificate.Successor.PolicyResult ||
		certificate.Predecessor.AuthorizationDigest != certificate.Successor.AuthorizationDigest {
		return validation(ValidationAxisMismatch, "predecessor and successor executions do not prove equal protected results")
	}
	preAxes := certificate.Predecessor.IdentityAxes()
	succAxes := certificate.Successor.IdentityAxes()
	if certificate.PredecessorAxes != preAxes || certificate.SuccessorAxes != succAxes {
		return validation(ValidationTampered, "certificate identity axes do not match embedded executions")
	}
	if preAxes.CompilerImplementation == succAxes.CompilerImplementation {
		return validation(ValidationAxisMismatch, "bounded successor certificate does not describe a different implementation")
	}
	for index, value := range preAxes.Values() {
		if index != 1 && value != succAxes.Values()[index] {
			return validation(ValidationAxisMismatch, fmt.Sprintf("identity axis %d differs outside compiler implementation", index+1))
		}
	}
	if certificate.GeneratedOutputDigest != certificate.Successor.GeneratedOutputDigest || certificate.GeneratedManifestDigest != certificate.Successor.GeneratedManifestDigest ||
		!reflect.DeepEqual(certificate.GeneratedSource, certificate.Successor.GeneratedSource) || !reflect.DeepEqual(certificate.GeneratedManifest, certificate.Successor.GeneratedManifest) ||
		!certificate.GeneratedBytesEqual || !certificate.GeneratedManifestEqual || !certificate.NormalizedSemanticEqual || !certificate.PolicyResultEqual || !certificate.FullTestContractEqual ||
		certificate.StrictPredecessorConsumption.Decision != DecisionRefuted || certificate.StrictPredecessorConsumption.State != "STRICT_IMPLEMENTATION_MISMATCH" {
		return validation(ValidationAxisMismatch, "compatibility proof does not establish the bounded implementation-only transition")
	}
	return nil
}

func (certificate Certificate) TransitiveCompatibility() bool {
	return certificate.Authorization.TransitiveCompatibility
}

func BuildAuthorization(candidateID, subjectDigest, successorCompilerDigest string) (Authorization, error) {
	authorization := Authorization{Schema: AuthorizationSchema, Mode: AuthorizationMode, CandidateStableID: candidateID,
		SubjectDigest: subjectDigest, SuccessorCompilerDigest: successorCompilerDigest, ScopeBounded: true,
		Scope: []ScopeSubject{{CandidateStableID: candidateID, SubjectDigest: subjectDigest}}, Authorized: true,
		TransitiveCompatibility: false}
	var err error
	authorization.AuthorizationID, err = authorization.ContentDigest()
	if err != nil {
		return Authorization{}, err
	}
	return authorization, ValidateAuthorization(authorization)
}

func BuildCertificate(predecessor, successor ExecutionReceipt, authorization Authorization, policy compatibilitypolicy.Policy) (Certificate, error) {
	if err := ValidateExecutionReceipt(predecessor); err != nil {
		return Certificate{}, err
	}
	if err := ValidateExecutionReceipt(successor); err != nil {
		return Certificate{}, err
	}
	if err := ValidateAuthorization(authorization); err != nil {
		return Certificate{}, err
	}
	if predecessor.CandidateStableID != successor.CandidateStableID || predecessor.SubjectDigest != successor.SubjectDigest || predecessor.SourceDigest != successor.SourceDigest ||
		predecessor.PolicyDigest != policy.SourceDigest || successor.PolicyDigest != policy.SourceDigest || predecessor.PolicyEvaluatorDigest != policy.EvaluatorDigest || successor.PolicyEvaluatorDigest != policy.EvaluatorDigest {
		return Certificate{}, validation(ValidationAxisMismatch, "execution receipts are not bound to the canonical policy and subject")
	}
	preAxes, succAxes := predecessor.IdentityAxes(), successor.IdentityAxes()
	if preAxes.CompilerImplementation == succAxes.CompilerImplementation {
		return Certificate{}, validation(ValidationAxisMismatch, "successor implementation identity is not different")
	}
	for index, value := range preAxes.Values() {
		if index != 1 && value != succAxes.Values()[index] {
			return Certificate{}, validation(ValidationAxisMismatch, "compatibility is broader than implementation-only")
		}
	}
	preDigest := receiptDigest(predecessor)
	succDigest := receiptDigest(successor)
	certificate := Certificate{Schema: CertificateSchema, Mode: CertificateMode, CandidateStableID: predecessor.CandidateStableID,
		SubjectDigest: predecessor.SubjectDigest, SourceDigest: predecessor.SourceDigest, PolicyDigest: policy.SourceDigest,
		PolicyEvaluatorDigest: policy.EvaluatorDigest, PredecessorReceiptDigest: preDigest, SuccessorReceiptDigest: succDigest,
		Predecessor: predecessor, Successor: successor, PredecessorAxes: preAxes, SuccessorAxes: succAxes,
		Authorization: authorization, AuthorizationDigest: authorization.AuthorizationID, ScopeBounded: true,
		Scope:                       []ScopeSubject{{CandidateStableID: predecessor.CandidateStableID, SubjectDigest: predecessor.SubjectDigest}},
		IndependentReplayExecutions: 2, GeneratedBytesEqual: string(predecessor.GeneratedSource) == string(successor.GeneratedSource),
		GeneratedManifestEqual:  string(predecessor.GeneratedManifest) == string(successor.GeneratedManifest),
		NormalizedSemanticEqual: predecessor.SemanticIRDigest == successor.SemanticIRDigest, PolicyResultEqual: predecessor.PolicyResult == successor.PolicyResult,
		FullTestContractEqual: predecessor.TestContractDigest == successor.TestContractDigest && predecessor.TestContractResult == successor.TestContractResult,
		GeneratedSource:       append([]byte(nil), successor.GeneratedSource...), GeneratedManifest: append([]byte(nil), successor.GeneratedManifest...),
		GeneratedOutputDigest: successor.GeneratedOutputDigest, GeneratedManifestDigest: successor.GeneratedManifestDigest,
		StrictPredecessorConsumption: StrictConsumption{Decision: DecisionRefuted, State: "STRICT_IMPLEMENTATION_MISMATCH", Reason: "COMPILER_IMPLEMENTATION_IDENTITY_MISMATCH"},
		RepositoryWrites:             0, LocalTestExecutions: 0}
	var err error
	certificate.CertificateID, err = certificate.ContentDigest()
	if err != nil {
		return Certificate{}, err
	}
	return certificate, ValidateCertificate(certificate)
}

func receiptDigest(receipt ExecutionReceipt) string {
	digest, _ := receipt.ContentDigest()
	return digest
}
