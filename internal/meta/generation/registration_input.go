package generation

import (
	"encoding/json"
	"io/fs"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

type RegistrationInputFailure struct {
	IndicatorID   string   `json:"indicator_id,omitempty"`
	RequestDigest string   `json:"request_digest"`
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func registrationDigestKnown(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validDigest(strings.TrimPrefix(value, "sha256:"))
}

func registrationRequestKnown(request syntaxregistration.Request) bool {
	if request.BaseVersion < 22 || request.BaseVersion > 10000 || request.Toolchain == "" ||
		!registrationDigestKnown(request.SnapshotDigest) || !registrationDigestKnown(request.SourceDigest) ||
		!fs.ValidPath(request.Case.Path) || !strings.HasPrefix(request.Case.Path, "examples/") ||
		!strings.HasSuffix(request.Case.Path, ".gooo") || request.Case.ID == "" ||
		string(request.Case.Kind) != "VALID" || string(request.Case.ExpectedDecision) != "PASS" ||
		request.Case.ExpectedDiagnostic != "" || string(request.Case.Scope) != "LANGUAGE_CAPABILITY" ||
		request.Case.MetaOperation != "replay-language-syntax" ||
		(request.Case.EntityFields && request.Case.ImplicitActivityPorts) {
		return false
	}
	if request.Case.ProofChoice != "FOUNDATION" && request.Case.ProofChoice != "COHERENCE" &&
		request.Case.ProofChoice != "REGRESSION" {
		return false
	}
	identity := request.ExecutionIdentity
	return identity.GoVersion == request.Toolchain && identity.GOOS != "" && identity.GOARCH != "" &&
		registrationDigestKnown(identity.ExecutableDigest) && registrationDigestKnown(identity.GoCommandDigest) &&
		registrationDigestKnown(identity.CompilerDigest)
}

func cloneRegistrationRequest(request *syntaxregistration.Request) *syntaxregistration.Request {
	if request == nil {
		return nil
	}
	raw, err := json.Marshal(request)
	if err != nil {
		return nil
	}
	cloned, err := syntaxregistration.DecodeRequest(raw)
	if err != nil {
		return nil
	}
	return &cloned
}

func ValidRegistrationActionInput(action Action) bool {
	if action.Operation != sourcepolicy.OperationRegisterSyntax {
		return action.RegistrationRequest == nil && action.SourceIndicator.OperationInputDigest == ""
	}
	request := action.RegistrationRequest
	return request != nil && registrationRequestKnown(*request) &&
		action.Subject == request.Case.Path && action.SubjectKind == sourcepolicy.SubjectKindRegistrationRequest &&
		action.MetricID == sourcepolicy.DimensionSyntaxRegistration &&
		action.SourceIndicator.OperationInputDigest == syntaxregistration.RequestDigest(*request) &&
		validRegistrationIndicator(action.SourceIndicator)
}

func validRegistrationIndicator(indicator sourcepolicy.Indicator) bool {
	binding, err := syntaxregistration.NativeBinding()
	if err != nil {
		return false
	}
	return indicator.Operation == sourcepolicy.OperationRegisterSyntax &&
		indicator.Role == sourcepolicy.IndicatorRoleDriver &&
		indicator.Producer == binding.InputActivityID && indicator.Consumer == binding.ActivityID &&
		indicator.MetricID == sourcepolicy.DimensionSyntaxRegistration &&
		indicator.SubjectKind == sourcepolicy.SubjectKindRegistrationRequest &&
		registrationDigestKnown(indicator.OperationInputDigest) &&
		indicator.Value >= 0 && indicator.Value <= 1 && indicator.Limit == 1 &&
		indicator.Relation == sourcepolicy.RelationEqual &&
		indicator.Satisfied == (indicator.Value == 1) && !indicator.Blocking &&
		indicator.Proof == sourcepolicy.ProofCoherence
}

// ObserveRegistrationInput connects an exact corpus fact to the native Gooo
// producer/consumer, typed request digest and common operation selector.
func ObserveRegistrationInput(repository fs.FS, request syntaxregistration.Request) (sourcepolicy.Indicator, error) {
	present, err := syntaxregistration.ObservePresence(repository, request)
	if err != nil {
		return sourcepolicy.Indicator{}, err
	}
	evidence, err := syntaxregistration.NativeBinding()
	if err != nil {
		return sourcepolicy.Indicator{}, err
	}
	value := 0
	if present {
		value = 1
	}
	return sourcepolicy.Indicator{MetricID: sourcepolicy.DimensionSyntaxRegistration,
		Family: sourcepolicy.FamilyConformance, Subject: request.Case.Path,
		SubjectKind: sourcepolicy.SubjectKindRegistrationRequest, Value: value, Limit: 1,
		Relation: sourcepolicy.RelationEqual, Applicability: sourcepolicy.ApplicabilityApplicable,
		ApplicabilityRule: sourcepolicy.ApplicabilityRuleDefault,
		ApplicabilityReason: sourcepolicy.ApplicabilityReasonCatalogApplicable,
		Satisfied: present, Role: sourcepolicy.IndicatorRoleDriver, Proof: sourcepolicy.ProofCoherence,
		Producer: evidence.InputActivityID, Consumer: evidence.ActivityID,
		Operation: sourcepolicy.OperationRegisterSyntax, OperationInputDigest: syntaxregistration.RequestDigest(request)}, nil
}
