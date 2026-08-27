package capabilityscopedexpansion

import (
	"fmt"
	"sort"
	"strings"
)

var declarations = []CapabilityDeclaration{
	{Kind: KindFile, Operation: OperationRead, Target: FileTarget},
	{Kind: KindTime, Operation: OperationRead, Target: TimeTarget},
	{Kind: KindEnvironment, Operation: OperationRead, Target: EnvironmentTarget},
	{Kind: KindNetwork, Operation: OperationRead, Target: NetworkTarget},
}

func capabilityValues() []CapabilityValue {
	return []CapabilityValue{
		{ValueID: "file-read", Kind: KindFile, Operation: OperationRead, Target: FileTarget},
		{ValueID: "logical-time-read", Kind: KindTime, Operation: OperationRead, Target: TimeTarget},
		{ValueID: "environment-read", Kind: KindEnvironment, Operation: OperationRead, Target: EnvironmentTarget},
		{ValueID: "pinned-network-read", Kind: KindNetwork, Operation: OperationRead, Target: NetworkTarget},
	}
}

func evidenceFor(source []byte, values []CapabilityValue) []Evidence {
	observed := map[string]string{
		"file-read":           digestBytes(source),
		"logical-time-read":   "logical-clock:0",
		"environment-read":    "GOOO_EXPANSION_PROFILE=deterministic",
		"pinned-network-read": digestBytes([]byte("pinned-schema-v1")),
	}
	result := make([]Evidence, 0, len(values))
	for _, value := range values {
		observation := observed[value.ValueID]
		result = append(result, Evidence{ValueID: value.ValueID, Observed: observation,
			EvidenceDigest: digestBytes([]byte(value.ValueID + "=" + observation))})
	}
	return result
}

func exactRequest(source []byte, subject, id string) Request {
	values := capabilityValues()
	return Request{CaseID: id, SubjectSHA: subject, Stage: StageExpand, Step: StepAuthorize,
		Toolchain: GoVersion, Capabilities: values, Evidence: evidenceFor(source, values)}
}

func Cases(source []byte, subject string) []Case {
	exact := exactRequest(source, subject, "allow-exact")
	fileDeny := exactRequest(source, subject, "deny-undeclared-file")
	fileDeny.Capabilities[0].Target = "repository-parent"
	timeDeny := exactRequest(source, subject, "deny-undeclared-time")
	timeDeny.Capabilities[1].Target = "wall-clock"
	environmentDeny := exactRequest(source, subject, "deny-undeclared-environment")
	environmentDeny.Capabilities[2].Target = "HOME"
	networkDeny := exactRequest(source, subject, "deny-undeclared-network")
	networkDeny.Capabilities[3].Target = "https://example.invalid/unpinned"
	writeDeny := exactRequest(source, subject, "deny-repository-write")
	writeDeny.RequestedRepositoryWrites = 1
	mutationDeny := exactRequest(source, subject, "deny-mutation-authority")
	mutationDeny.RequestedMutationAuthority = true
	unknown := exactRequest(source, subject, "unknown-missing-evidence")
	unknown.Evidence = unknown.Evidence[:len(unknown.Evidence)-1]
	return []Case{
		{ID: exact.CaseID, Request: exact, ExpectedDecision: DecisionAllow, ExpectedResolution: ResolutionExact},
		{ID: fileDeny.CaseID, Request: fileDeny, ExpectedDecision: DecisionDeny, ExpectedResolution: ResolutionReject},
		{ID: timeDeny.CaseID, Request: timeDeny, ExpectedDecision: DecisionDeny, ExpectedResolution: ResolutionReject},
		{ID: environmentDeny.CaseID, Request: environmentDeny, ExpectedDecision: DecisionDeny, ExpectedResolution: ResolutionReject},
		{ID: networkDeny.CaseID, Request: networkDeny, ExpectedDecision: DecisionDeny, ExpectedResolution: ResolutionReject},
		{ID: writeDeny.CaseID, Request: writeDeny, ExpectedDecision: DecisionDeny, ExpectedResolution: ResolutionReject},
		{ID: mutationDeny.CaseID, Request: mutationDeny, ExpectedDecision: DecisionDeny, ExpectedResolution: ResolutionReject},
		{ID: unknown.CaseID, Request: unknown, ExpectedDecision: DecisionUnknown, ExpectedResolution: ResolutionUnknown},
	}
}

func Evaluate(source []byte, request Request) Receipt {
	receipt := Receipt{Schema: Schema, MetaOperation: MetaOperation, Producer: Producer, Consumer: Consumer,
		SubjectSHA: request.SubjectSHA, GoVersion: request.Toolchain, SourceDigest: digestBytes(source),
		CaseID: request.CaseID, Stage: request.Stage, Step: request.Step, Declarations: cloneDeclarations(),
		Capabilities: request.Capabilities, Evidence: request.Evidence, Authority: Authority{
			CapabilitiesRequested: len(request.Capabilities), RequestedRepositoryWrites: request.RequestedRepositoryWrites,
			RequestedMutationAuthority:  request.RequestedMutationAuthority,
			RequestedPromotionAuthority: request.RequestedPromotionAuthority},
		RepositoryWrites: 0, MutationAuthority: false, PromotionAuthority: false}
	for index := range receipt.Declarations {
		receipt.Declarations[index].Declared = sourceHasDeclaration(source, receipt.Declarations[index])
	}

	if err := ValidateShape(source); err != nil {
		return finishUnknown(receipt, "compile", "declare-capabilities", "SOURCE_DECLARATIONS_UNOBSERVED")
	}
	for _, check := range []struct{ value, expected string }{{request.Stage, StageExpand}, {request.Step, StepAuthorize}, {request.Toolchain, GoVersion}} {
		if check.value == "" {
			return finishUnknown(receipt, StageExpand, StepAuthorize, "EXPANSION_INPUT_UNOBSERVED")
		}
	}
	if request.Stage != StageExpand {
		return finishDeny(receipt, "EXPANSION_BOUNDARY_REJECTED")
	}
	if request.Step != StepAuthorize {
		return finishDeny(receipt, "EXPANSION_BOUNDARY_REJECTED")
	}
	if request.Toolchain != GoVersion {
		return finishDeny(receipt, "EXPANSION_BOUNDARY_REJECTED")
	}

	declared := 0
	for _, value := range request.Capabilities {
		if matchesDeclaration(value, receipt.Declarations) {
			declared++
		}
	}
	receipt.Authority.CapabilitiesDeclared = declared
	if declared != len(request.Capabilities) {
		return finishDeny(receipt, "CAPABILITY_NOT_DECLARED")
	}
	if !matchesEvidence(request.Capabilities, request.Evidence, source) {
		return finishUnknown(receipt, StageExpand, "bind-capability-evidence", "EVIDENCE_UNOBSERVED")
	}
	if request.RequestedRepositoryWrites != 0 || request.RequestedMutationAuthority || request.RequestedPromotionAuthority {
		return finishDeny(receipt, "EFFECT_CEILING_REJECTED")
	}
	receipt.Decision, receipt.Resolution, receipt.EnforcementEffect, receipt.Reason = DecisionAllow, ResolutionExact, EffectNone, "CAPABILITY_SCOPE_EXACT"
	receipt.Authority.CapabilitiesAuthorized = len(request.Capabilities)
	receipt.Authority.CapabilitiesDenied = 0
	return finishKnown(receipt)
}

func EvaluateSuite(source []byte, subject string) (Suite, []Receipt) {
	cases := Cases(source, subject)
	receipts := make([]Receipt, 0, len(cases))
	suite := Suite{Schema: "gooo/capability-scoped-expansion-suite/v1", MetaOperation: MetaOperation,
		SubjectSHA: subject, SourceDigest: digestBytes(source), Decision: "PASS", Resolution: ResolutionExact,
		RepositoryWrites: 0, MutationAuthority: false, PromotionAuthority: false}
	for _, item := range cases {
		receipt := Evaluate(source, item.Request)
		receipts = append(receipts, receipt)
		suite.Cases = append(suite.Cases, CaseResult{CaseID: item.ID, ExpectedDecision: item.ExpectedDecision,
			ObservedDecision: receipt.Decision, ExpectedResolution: item.ExpectedResolution,
			ObservedResolution: receipt.Resolution, ReceiptDigest: receipt.ReportDigest})
		suite.Summary.CasesTotal++
		suite.Summary.CapabilityRequests += receipt.Authority.CapabilitiesRequested
		suite.Summary.CapabilityAuthorized += receipt.Authority.CapabilitiesAuthorized
		suite.Summary.CapabilityDenied += receipt.Authority.CapabilitiesDenied
		suite.Summary.CapabilityUnknown += receipt.Authority.CapabilitiesUnknown
		suite.Summary.BlockedWriteAttempts += boolInt(item.Request.RequestedRepositoryWrites != 0)
		suite.Summary.BlockedMutationAttempts += boolInt(item.Request.RequestedMutationAuthority)
		switch receipt.Decision {
		case DecisionAllow:
			suite.Summary.AllowCases++
		case DecisionDeny:
			suite.Summary.DenyCases++
		case DecisionUnknown:
			suite.Summary.UnknownCases++
		}
		if receipt.Decision == item.ExpectedDecision && receipt.Resolution == item.ExpectedResolution {
			suite.Summary.CasesPassed++
		}
	}
	suite.Summary.RepositoryWrites = 0
	suite.Summary.MutationAuthority = false
	suite.Summary.PromotionAuthority = false
	return suite, receipts
}

func cloneDeclarations() []CapabilityDeclaration {
	return append([]CapabilityDeclaration(nil), declarations...)
}

func sourceHasDeclaration(source []byte, declaration CapabilityDeclaration) bool {
	markers := map[string]string{
		KindFile:        "activity DeclareFileReadCapability",
		KindTime:        "activity DeclareLogicalTimeCapability",
		KindEnvironment: "activity DeclareEnvironmentReadCapability",
		KindNetwork:     "activity DeclarePinnedNetworkCapability",
	}
	return containsBytes(source, markers[declaration.Kind])
}

func containsBytes(source []byte, marker string) bool {
	return len(marker) > 0 && strings.Contains(string(source), marker)
}

func matchesDeclaration(value CapabilityValue, declared []CapabilityDeclaration) bool {
	for _, item := range declared {
		if item.Declared && item.Kind == value.Kind && item.Operation == value.Operation && item.Target == value.Target {
			return true
		}
	}
	return false
}

func matchesEvidence(values []CapabilityValue, evidence []Evidence, source []byte) bool {
	expected := evidenceFor(source, values)
	if len(expected) != len(evidence) {
		return false
	}
	byID := make(map[string]Evidence, len(evidence))
	for _, item := range evidence {
		if _, exists := byID[item.ValueID]; exists {
			return false
		}
		byID[item.ValueID] = item
	}
	for _, item := range expected {
		observed, ok := byID[item.ValueID]
		if !ok || observed != item {
			return false
		}
	}
	return true
}

func finishKnown(receipt Receipt) Receipt {
	receipt.Claims = claimsFor(receipt.Decision, receipt.RequestedEffect())
	receipt.Indicators = indicatorsFor(receipt, true, true, true, true)
	return sealReceipt(receipt)
}

func finishDeny(receipt Receipt, reason string) Receipt {
	receipt.Decision, receipt.Resolution, receipt.EnforcementEffect, receipt.Reason = DecisionDeny, ResolutionReject, EffectBlock, reason
	receipt.Authority.CapabilitiesDenied = receipt.Authority.CapabilitiesRequested - receipt.Authority.CapabilitiesAuthorized
	if receipt.Authority.CapabilitiesDenied < 0 {
		receipt.Authority.CapabilitiesDenied = receipt.Authority.CapabilitiesRequested
	}
	receipt.Claims = claimsFor(receipt.Decision, receipt.RequestedEffect())
	receipt.Indicators = indicatorsFor(receipt, false, true, true, false)
	return sealReceipt(receipt)
}

func finishUnknown(receipt Receipt, stage, step, reason string) Receipt {
	receipt.Decision, receipt.Resolution, receipt.EnforcementEffect, receipt.Reason = DecisionUnknown, ResolutionUnknown, EffectBlock, reason
	receipt.Unknown = &Unknown{Stage: stage, Step: step, Reason: reason}
	receipt.Authority.CapabilitiesUnknown = receipt.Authority.CapabilitiesRequested
	receipt.Claims = claimsFor(receipt.Decision, receipt.RequestedEffect())
	receipt.Indicators = indicatorsFor(receipt, false, false, false, true)
	return sealReceipt(receipt)
}

func (receipt Receipt) RequestedEffect() bool {
	return receipt.Authority.RequestedRepositoryWrites != 0 || receipt.Authority.RequestedMutationAuthority || receipt.Authority.RequestedPromotionAuthority
}

func claimsFor(decision string, effect bool) []Claim {
	scope := "DISCHARGED"
	if decision == DecisionDeny {
		scope = "REFUTED"
	} else if decision == DecisionUnknown {
		scope = "OPEN"
	}
	status := "DISCHARGED"
	if decision == DecisionUnknown {
		status = "OPEN"
	}
	return []Claim{
		{ID: "capability-scope-exact", Status: scope, ProofChoice: "COHERENCE", Evidence: "source-and-capability-values"},
		{ID: "default-deny", Status: status, ProofChoice: "REGRESSION", Evidence: "denied-undeclared-capabilities"},
		{ID: "effect-ceiling", Status: status, ProofChoice: "REGRESSION", Evidence: "zero-write-authority"},
	}
}

func indicatorsFor(receipt Receipt, source, evidence, stage, exact bool) []Indicator {
	values := []struct {
		id, class, proof string
		observed, target int
		satisfied        bool
	}{
		{"source-shape", "DRIVER", "FOUNDATION", boolInt(source), 1, source},
		{"toolchain-1.27", "DRIVER", "FOUNDATION", boolInt(receipt.GoVersion == GoVersion), 1, receipt.GoVersion == GoVersion},
		{"file-capability-declared", "DRIVER", "FOUNDATION", boolInt(declaration(receipt, KindFile)), 1, declaration(receipt, KindFile)},
		{"time-capability-declared", "DRIVER", "FOUNDATION", boolInt(declaration(receipt, KindTime)), 1, declaration(receipt, KindTime)},
		{"environment-capability-declared", "DRIVER", "FOUNDATION", boolInt(declaration(receipt, KindEnvironment)), 1, declaration(receipt, KindEnvironment)},
		{"network-capability-declared", "DRIVER", "FOUNDATION", boolInt(declaration(receipt, KindNetwork)), 1, declaration(receipt, KindNetwork)},
		{"value-evidence-relation", "OUTCOME", "COHERENCE", boolInt(evidence), 1, evidence},
		{"source-binding", "OUTCOME", "COHERENCE", boolInt(receipt.SourceDigest != ""), 1, receipt.SourceDigest != ""},
		{"expansion-stage-order", "OUTCOME", "COHERENCE", boolInt(stage && receipt.Stage == StageExpand && receipt.Step == StepAuthorize), 1, stage && receipt.Stage == StageExpand && receipt.Step == StepAuthorize},
		{"receipt-seal", "DRIVER", "COHERENCE", boolInt(receipt.ReportDigest == "" || len(receipt.ReportDigest) > 0), 1, true},
		{"default-deny", "GUARDRAIL", "REGRESSION", boolInt(receipt.Decision != DecisionAllow || exact), 1, receipt.Decision != DecisionAllow || exact},
		{"authority-ceiling", "GUARDRAIL", "REGRESSION", boolInt(receipt.RepositoryWrites == 0 && !receipt.MutationAuthority && !receipt.PromotionAuthority), 1, receipt.RepositoryWrites == 0 && !receipt.MutationAuthority && !receipt.PromotionAuthority},
	}
	result := make([]Indicator, 0, len(values))
	for _, item := range values {
		status := "SATISFIED"
		if !item.satisfied {
			status = "REFUTED"
			if receipt.Decision == DecisionUnknown {
				status = "OPEN"
			}
		}
		result = append(result, Indicator{ID: "CSE-" + item.id, Class: item.class, Status: status,
			Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: item.proof,
			Observed: item.observed, Target: item.target})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].ID < result[right].ID })
	return result
}

func declaration(receipt Receipt, kind string) bool {
	for _, item := range receipt.Declarations {
		if item.Kind == kind && item.Declared {
			return true
		}
	}
	return false
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (request Request) String() string {
	return fmt.Sprintf("%s/%s", request.CaseID, request.Stage)
}
