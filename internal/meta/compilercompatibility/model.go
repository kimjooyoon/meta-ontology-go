package compilercompatibility

import (
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const (
	CertificateSchema           = "gooo/compiler-successor-compatibility-certificate/v1"
	CertificateMode             = "caller-owned-immutable-successor-compatibility"
	AuthorizationSchema         = "gooo/compiler-successor-compatibility-authorization/v1"
	AuthorizationMode           = "explicit-caller-owned-successor-scope"
	ReportSchema                = "gooo/compiler-successor-compatibility-report/v1"
	ConsumptionSchema           = "gooo/compiler-successor-compatibility-consumption/v1"
	CanonicalPolicyPath         = "examples/self-improvement-observation/observation.gooo"
	UnsupportedFrontierDecision = compatibilitypolicy.DecisionUnknown

	DecisionClosed  = compatibilitypolicy.DecisionClosed
	DecisionUnknown = compatibilitypolicy.DecisionUnknown
	DecisionRefuted = compatibilitypolicy.DecisionRefuted

	ReasonExactStrictReplay      = "EXACT_STRICT_REPLAY"
	ReasonBoundedSuccessorReplay = "BOUNDED_IMPLEMENTATION_SUCCESSOR_REPLAY"
	ReasonMissingSuccessorReplay = "MISSING_SUCCESSOR_REPLAY"
	ReasonUnboundedScope         = "UNBOUNDED_COMPATIBILITY_SCOPE"
	ReasonAxisMismatch           = "COMPATIBILITY_AXIS_MISMATCH"
	ReasonTamperedCertificate    = "TAMPERED_COMPATIBILITY_CERTIFICATE"
	ReasonWidenedScope           = "WIDENED_COMPATIBILITY_SCOPE"
	ReasonMissingCertificate     = "MISSING_COMPATIBILITY_CERTIFICATE"
	ReasonAmbiguousEvidence      = "AMBIGUOUS_COMPATIBILITY_EVIDENCE"
	ReasonCurrentSubjectMismatch = "CURRENT_SUBJECT_MISMATCH"
)

var UnsupportedFrontierClaims = []string{"SIGNATURES", "EXPIRY", "REVOCATION", "TRUST_ROOT"}

// IdentityAxes are the compatibility boundary. Only the compiler
// implementation identity may differ in a bounded successor certificate.
type IdentityAxes struct {
	SemanticIdentity          string `json:"semantic_identity"`
	CompilerImplementation    string `json:"compiler_implementation_identity"`
	GoToolchain               string `json:"go_toolchain_identity"`
	PolicyIdentity            string `json:"policy_identity"`
	GeneratedArtifactIdentity string `json:"generated_artifact_identity"`
	TestContractIdentity      string `json:"test_contract_identity"`
	AuthorizationIdentity     string `json:"authorization_identity"`
}

func (axes IdentityAxes) Values() []string {
	return []string{axes.SemanticIdentity, axes.CompilerImplementation, axes.GoToolchain, axes.PolicyIdentity,
		axes.GeneratedArtifactIdentity, axes.TestContractIdentity, axes.AuthorizationIdentity}
}

func (axes IdentityAxes) Complete() bool {
	for _, value := range axes.Values() {
		if !cache.Digest(value).Known() {
			return false
		}
	}
	return true
}

type ExecutionReceipt struct {
	Schema                       string `json:"schema"`
	Role                         string `json:"role"`
	CandidateStableID            string `json:"candidate_stable_id"`
	SubjectDigest                string `json:"subject_digest"`
	SourceDigest                 string `json:"source_digest"`
	SemanticIRDigest             string `json:"semantic_ir_digest"`
	GeneratedOutputDigest        string `json:"generated_output_digest"`
	GeneratedManifestDigest      string `json:"generated_manifest_digest"`
	GeneratedSource              []byte `json:"generated_source"`
	GeneratedManifest            []byte `json:"generated_manifest"`
	PolicyDigest                 string `json:"policy_digest"`
	PolicyEvaluatorDigest        string `json:"policy_evaluator_digest"`
	PolicyResult                 string `json:"policy_result"`
	CompilerImplementationDigest string `json:"compiler_implementation_digest"`
	GoToolchainDigest            string `json:"go_toolchain_digest"`
	TestContractDigest           string `json:"test_contract_digest"`
	TestContractResult           string `json:"test_contract_result"`
	AuthorizationDigest          string `json:"authorization_digest"`
}

func (receipt ExecutionReceipt) IdentityAxes() IdentityAxes {
	policyIdentity := cache.HashBytes([]byte(receipt.PolicyDigest + "\x00" + receipt.PolicyEvaluatorDigest + "\x00" + receipt.PolicyResult)).String()
	artifactIdentity := cache.HashBytes([]byte(receipt.GeneratedOutputDigest + "\x00" + receipt.GeneratedManifestDigest)).String()
	return IdentityAxes{
		SemanticIdentity:          receipt.SemanticIRDigest,
		CompilerImplementation:    receipt.CompilerImplementationDigest,
		GoToolchain:               receipt.GoToolchainDigest,
		PolicyIdentity:            policyIdentity,
		GeneratedArtifactIdentity: artifactIdentity,
		TestContractIdentity:      receipt.TestContractDigest,
		AuthorizationIdentity:     receipt.AuthorizationDigest,
	}
}

func (receipt ExecutionReceipt) ContentDigest() (string, error) {
	copy := receipt
	copy.Schema = ""
	copy.GeneratedSource = append([]byte(nil), receipt.GeneratedSource...)
	copy.GeneratedManifest = append([]byte(nil), receipt.GeneratedManifest...)
	return digestOf(copy)
}

type ScopeSubject struct {
	CandidateStableID string `json:"candidate_stable_id"`
	SubjectDigest     string `json:"subject_digest"`
}

type Authorization struct {
	Schema                  string         `json:"schema"`
	AuthorizationID         string         `json:"authorization_id"`
	Mode                    string         `json:"mode"`
	CandidateStableID       string         `json:"candidate_stable_id"`
	SubjectDigest           string         `json:"subject_digest"`
	SuccessorCompilerDigest string         `json:"successor_compiler_digest"`
	ScopeBounded            bool           `json:"scope_bounded"`
	Scope                   []ScopeSubject `json:"scope"`
	Authorized              bool           `json:"authorized"`
	TransitiveCompatibility bool           `json:"transitive_compatibility"`
}

func (authorization Authorization) ContentDigest() (string, error) {
	copy := authorization
	copy.AuthorizationID = ""
	return digestOf(copy)
}

type StrictConsumption struct {
	Decision string `json:"decision"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
}

type Certificate struct {
	Schema                       string            `json:"schema"`
	CertificateID                string            `json:"certificate_id"`
	Mode                         string            `json:"mode"`
	CandidateStableID            string            `json:"candidate_stable_id"`
	SubjectDigest                string            `json:"subject_digest"`
	SourceDigest                 string            `json:"source_digest"`
	PolicyDigest                 string            `json:"policy_digest"`
	PolicyEvaluatorDigest        string            `json:"policy_evaluator_digest"`
	PredecessorReceiptDigest     string            `json:"predecessor_receipt_digest"`
	SuccessorReceiptDigest       string            `json:"successor_receipt_digest"`
	Predecessor                  ExecutionReceipt  `json:"predecessor"`
	Successor                    ExecutionReceipt  `json:"successor"`
	PredecessorAxes              IdentityAxes      `json:"predecessor_axes"`
	SuccessorAxes                IdentityAxes      `json:"successor_axes"`
	Authorization                Authorization     `json:"authorization"`
	AuthorizationDigest          string            `json:"authorization_digest"`
	ScopeBounded                 bool              `json:"scope_bounded"`
	Scope                        []ScopeSubject    `json:"scope"`
	IndependentReplayExecutions  int               `json:"independent_replay_executions"`
	GeneratedBytesEqual          bool              `json:"generated_bytes_equal"`
	GeneratedManifestEqual       bool              `json:"generated_manifest_equal"`
	NormalizedSemanticEqual      bool              `json:"normalized_semantic_equal"`
	PolicyResultEqual            bool              `json:"policy_result_equal"`
	FullTestContractEqual        bool              `json:"full_test_contract_equal"`
	GeneratedSource              []byte            `json:"generated_source"`
	GeneratedManifest            []byte            `json:"generated_manifest"`
	GeneratedOutputDigest        string            `json:"generated_output_digest"`
	GeneratedManifestDigest      string            `json:"generated_manifest_digest"`
	StrictPredecessorConsumption StrictConsumption `json:"strict_predecessor_consumption"`
	RepositoryWrites             int               `json:"repository_writes"`
	LocalTestExecutions          int               `json:"local_test_executions"`
}

func (certificate Certificate) ContentDigest() (string, error) {
	copy := certificate
	copy.CertificateID = ""
	return digestOf(copy)
}

type UnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type AxisComparison struct {
	Axis        string `json:"axis"`
	Predecessor string `json:"predecessor"`
	Successor   string `json:"successor"`
	Equal       bool   `json:"equal"`
}

type Evaluation struct {
	Decision                    string           `json:"decision"`
	Reason                      string           `json:"reason"`
	Unknown                     *UnknownState    `json:"unknown"`
	Axes                        []AxisComparison `json:"axis_comparisons"`
	ExactSubjectBinding         bool             `json:"exact_subject_binding"`
	MismatchDetected            bool             `json:"mismatch_detected"`
	FallbackAttempted           bool             `json:"fallback_attempted"`
	FallbackRejected            bool             `json:"fallback_rejected"`
	IndependentReplayExecutions int              `json:"independent_replay_executions"`
	CompatibilityHit            bool             `json:"compatibility_hit"`
}

type ConsumptionReport struct {
	Schema                        string            `json:"schema"`
	Lifecycle                     string            `json:"lifecycle"`
	Decision                      string            `json:"decision"`
	Reason                        string            `json:"reason"`
	Unknown                       *UnknownState     `json:"unknown"`
	CertificateDigest             string            `json:"certificate_digest"`
	CandidateStableID             string            `json:"candidate_stable_id"`
	SubjectDigest                 string            `json:"subject_digest"`
	SourceDigest                  string            `json:"source_digest"`
	StrictPredecessorConsumption  StrictConsumption `json:"strict_predecessor_consumption"`
	IdentityAxisCount             int               `json:"identity_axis_count"`
	AxisComparisons               []AxisComparison  `json:"axis_comparisons"`
	ImplementationDigestEqual     bool              `json:"implementation_digest_equal"`
	ImplementationDigestDifferent bool              `json:"implementation_digest_different"`
	SemanticEqual                 bool              `json:"semantic_equal"`
	GeneratedBytesEqual           bool              `json:"generated_bytes_equal"`
	GeneratedManifestEqual        bool              `json:"generated_manifest_equal"`
	PolicyResultEqual             bool              `json:"policy_result_equal"`
	FullTestContractEqual         bool              `json:"full_test_contract_equal"`
	IndependentReplayExecutions   int               `json:"independent_replay_executions"`
	TestContractReplays           int               `json:"test_contract_replays"`
	CompatibilityScopeSubjects    int               `json:"compatibility_scope_subjects"`
	CertificateCount              int               `json:"certificate_count"`
	CertificateBytes              int               `json:"certificate_bytes"`
	CompatibilityHits             int               `json:"compatibility_hits"`
	CompatibilityMisses           int               `json:"compatibility_misses"`
	MismatchDetections            int               `json:"mismatch_detections"`
	CertificateTamperDetections   int               `json:"certificate_tamper_detections"`
	ScopeWideningDetections       int               `json:"scope_widening_detections"`
	FallbackAttempts              int               `json:"fallback_attempts"`
	FallbackAccepted              int               `json:"fallback_accepted"`
	FallbackRejected              int               `json:"fallback_rejected"`
	EvidenceArtifactCount         int               `json:"evidence_artifact_count"`
	ContinuityEdgeCount           int               `json:"continuity_edge_count"`
	Claim                         string            `json:"claim"`
	PerformanceClaim              bool              `json:"performance_claim"`
	GeneralCompatibilityClaim     bool              `json:"general_compiler_compatibility_claim"`
	UnsupportedFrontierDecision   string            `json:"unsupported_frontier_decision"`
	UnsupportedFrontierClaims     []string          `json:"unsupported_frontier_claims"`
	OutputFile                    string            `json:"output_file"`
	ManifestFile                  string            `json:"manifest_file"`
	ArtifactCount                 int               `json:"artifact_count"`
	OutputBytes                   int64             `json:"output_bytes"`
	RepositoryWrites              int               `json:"repository_writes"`
	LocalTestExecutions           int               `json:"local_test_executions"`
	WallMS                        int64             `json:"wall_ms"`
	PeakRSSKib                    int64             `json:"peak_rss_kib"`
}

type Request struct {
	Mode              string
	CandidateStableID string
	SubjectDigest     string
	SourceDigest      string
	Current           ExecutionReceipt
	Certificate       *Certificate
}

func digestOf(value any) (string, error) {
	digest, err := cache.DigestOf(value)
	if err != nil {
		return "", fmt.Errorf("content digest: %w", err)
	}
	return digest.String(), nil
}

func TestContractDigest() string {
	return cache.HashBytes([]byte(compatibilitypolicy.TestContract)).String()
}

func CurrentToolchainDigest() string { return generation.SemanticRetentionToolchainDigest() }

func validDigest(value string) bool { return cache.Digest(value).Known() }

var errMissingEvidence = errors.New("compatibility evidence is missing")
