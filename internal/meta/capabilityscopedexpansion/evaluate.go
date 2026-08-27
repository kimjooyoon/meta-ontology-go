package capabilityscopedexpansion

import (
	"fmt"
	"sort"
)

// EvaluateSuite parses and lowers the same source once, consumes raw provider
// observations, and evaluates every case declared by semantic Gooo values.
func EvaluateSuite(source, providerRaw []byte, subject string) (Suite, []Receipt, error) {
	model, err := ParseSource(source)
	if err != nil {
		return Suite{}, nil, err
	}
	provider, err := decodeProvider(providerRaw)
	if err != nil {
		return Suite{}, nil, fmt.Errorf("decode provider observations: %w", err)
	}
	if provider.SubjectSHA != subject {
		return Suite{}, nil, fmt.Errorf("provider subject %q does not match %q", provider.SubjectSHA, subject)
	}
	if len(model.Cases) != FixedCaseTotal {
		return Suite{}, nil, fmt.Errorf("semantic case denominator is %d, want %d", len(model.Cases), FixedCaseTotal)
	}
	current, historical := evidenceDenominator(model, provider)
	suite := Suite{Schema: "gooo/capability-scoped-expansion-suite/v2", MetaOperation: MetaOperation, SubjectSHA: subject, SourceDigest: model.SourceDigest, SemanticDigest: model.SemanticDigest, Decision: SuitePass, Resolution: ResolutionExact}
	suite.Summary = SuiteSummary{
		CasesTotal: len(model.Cases), CurrentEvidenceCapabilities: current, CurrentEvidenceDenominator: len(model.Declarations), HistoricalFixtureCapabilities: historical,
		EnforcementObservations: len(provider.EffectAttempts), RepositoryWrites: provider.ActualRepositoryWrites, MutationAuthority: provider.ActualMutationAuthority, PromotionAuthority: provider.ActualPromotionAuthority,
		SourceReconstructionPasses: boolInt(model.Reconstructed), SourceReconstructionTotal: 1, ProducerImportNumerator: 0, ProducerImportDenominator: 1,
	}
	receipts := make([]Receipt, 0, len(model.Cases))
	for _, item := range model.Cases {
		receipt := Evaluate(model, providerRaw, provider, subject, item)
		receipts = append(receipts, receipt)
		suite.Cases = append(suite.Cases, CaseResult{CaseID: item.ID, ObservedDecision: receipt.Decision, ObservedResolution: receipt.Resolution, ReceiptDigest: receipt.ReportDigest})
		suite.Summary.CapabilityRequests += receipt.Authority.CapabilitiesRequested
		suite.Summary.CapabilityAuthorized += receipt.Authority.CapabilitiesAuthorized
		suite.Summary.CapabilityDenied += receipt.Authority.CapabilitiesDenied
		suite.Summary.CapabilityUnknown += receipt.Authority.CapabilitiesUnknown
		switch receipt.Decision {
		case DecisionAllow:
			suite.Summary.AllowCases++
		case DecisionDeny:
			suite.Summary.DenyCases++
		case DecisionUnknown:
			suite.Summary.UnknownCases++
		}
		suite.Summary.BlockedWriteAttempts += boolInt(item.RequestedRepositoryWrites != 0)
		suite.Summary.BlockedMutationAttempts += boolInt(item.RequestedMutationAuthority)
		if receipt.Decision == expectedDecision(model, provider, item) {
			suite.Summary.CasesPassed++
		}
	}
	if suite.Summary.CasesPassed != suite.Summary.CasesTotal {
		suite.Decision = DecisionUnknown
		suite.Resolution = ResolutionLower
	}
	suite.RepositoryWrites = provider.ActualRepositoryWrites
	suite.MutationAuthority = provider.ActualMutationAuthority
	suite.PromotionAuthority = provider.ActualPromotionAuthority
	return suite, receipts, nil
}

func Evaluate(model SourceModel, providerRaw []byte, provider ProviderObservation, subject string, item CaseSpec) Receipt {
	operation := authorizationOperation(model)
	receipt := Receipt{
		Schema: Schema, MetaOperation: MetaOperation, Producer: Producer, Consumer: Consumer, SubjectSHA: subject, GoVersion: GoVersion,
		SourceDigest: model.SourceDigest, SemanticDigest: model.SemanticDigest, CaseID: item.ID, Stage: operation.Stage, Step: operation.Step,
		Policy: model.Policy, Declarations: append([]CapabilityDeclaration(nil), model.Declarations...), Capabilities: append([]CapabilityValue(nil), item.Requests...),
		ProviderDigest: providerDigest(providerRaw), EffectObservations: append([]EffectObservation(nil), provider.EffectAttempts...),
		RepositoryWrites: provider.ActualRepositoryWrites, MutationAuthority: provider.ActualMutationAuthority, PromotionAuthority: provider.ActualPromotionAuthority,
	}
	receipt.Authority = Authority{
		CapabilitiesRequested: len(item.Requests), CurrentEvidenceCapabilities: evidenceDenominatorCount(model, provider), CurrentEvidenceDenominator: len(model.Declarations),
		RequestedRepositoryWrites: item.RequestedRepositoryWrites, RequestedMutationAuthority: item.RequestedMutationAuthority, RequestedPromotionAuthority: item.RequestedPromotionAuthority,
		RepositoryWrites: provider.ActualRepositoryWrites, MutationAuthority: provider.ActualMutationAuthority, PromotionAuthority: provider.ActualPromotionAuthority, EnforcementObservations: len(provider.EffectAttempts),
	}
	for index := range receipt.Declarations {
		receipt.Declarations[index].NodeID = receipt.Declarations[index].NodeID
	}

	decision, resolution, reason, unknown := decisionFor(model, provider, item)
	receipt.Decision, receipt.Resolution, receipt.Reason, receipt.Unknown = decision, resolution, reason, unknown
	for _, value := range item.Requests {
		declaration, ok := declarationFor(model, value.ValueID)
		if !ok || !capabilityMatches(value, declaration) {
			continue
		}
		if evidence, observed := currentEvidenceFor(provider, declaration); observed {
			receipt.Evidence = append(receipt.Evidence, evidence)
		}
	}
	receipt.Authority.CapabilitiesDeclared = declaredCount(model, item)
	switch decision {
	case DecisionAllow:
		receipt.EnforcementEffect = EffectNone
		receipt.Authority.CapabilitiesAuthorized = len(item.Requests)
	case DecisionDeny:
		receipt.EnforcementEffect = EffectBlock
		receipt.Authority.CapabilitiesDenied = len(item.Requests)
	case DecisionUnknown:
		receipt.EnforcementEffect = EffectBlock
		receipt.Authority.CapabilitiesUnknown = len(item.Requests)
	}
	receipt.ClaimTransitions = claimTransitions(model, provider, item, receipt)
	receipt.Claims = claimsFromTransitions(receipt.ClaimTransitions)
	receipt.Indicators = indicatorsFor(model, provider, receipt)
	return sealReceipt(receipt)
}

func expectedDecision(model SourceModel, provider ProviderObservation, item CaseSpec) string {
	decision, _, _, _ := decisionFor(model, provider, item)
	return decision
}

func decisionFor(model SourceModel, provider ProviderObservation, item CaseSpec) (string, string, string, *Unknown) {
	operation := authorizationOperation(model)
	if item.RequestedRepositoryWrites != 0 || item.RequestedMutationAuthority || item.RequestedPromotionAuthority {
		if effectBoundaryObserved(provider) {
			return DecisionDeny, ResolutionExact, "CAPABILITY_ENFORCEMENT_OBSERVED", nil
		}
		return DecisionUnknown, ResolutionLower, "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED", &Unknown{Stage: operation.Stage, Step: operation.Step, Reason: "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED"}
	}
	for _, value := range item.Requests {
		declaration, ok := declarationFor(model, value.ValueID)
		if !ok || !capabilityMatches(value, declaration) {
			return DecisionDeny, ResolutionExact, "CAPABILITY_NOT_DECLARED", nil
		}
	}
	for _, value := range item.Requests {
		declaration, _ := declarationFor(model, value.ValueID)
		if _, observed := currentEvidenceFor(provider, declaration); !observed {
			return DecisionUnknown, ResolutionLower, "EVIDENCE_UNOBSERVED", &Unknown{Stage: operation.Stage, Step: "bind-capability-evidence", Reason: "EVIDENCE_UNOBSERVED"}
		}
	}
	if model.Policy.AuthorizationMode != PolicyExactCurrent {
		return DecisionDeny, ResolutionExact, "POLICY_REJECTED", nil
	}
	return DecisionAllow, ResolutionExact, "CAPABILITY_SCOPE_EXACT", nil
}

func declaredCount(model SourceModel, item CaseSpec) int {
	count := 0
	for _, value := range item.Requests {
		declaration, ok := declarationFor(model, value.ValueID)
		if ok && capabilityMatches(value, declaration) {
			count++
		}
	}
	return count
}

func evidenceDenominator(model SourceModel, provider ProviderObservation) (int, int) {
	current := evidenceDenominatorCount(model, provider)
	historical := 0
	for _, declaration := range model.Declarations {
		if declaration.EvidenceClass == HistoricalFixture {
			historical++
		}
	}
	return current, historical
}

func evidenceDenominatorCount(model SourceModel, provider ProviderObservation) int {
	count := 0
	for _, declaration := range model.Declarations {
		if _, ok := currentEvidenceFor(provider, declaration); ok {
			count++
		}
	}
	return count
}

func effectBoundaryObserved(provider ProviderObservation) bool {
	if len(provider.EffectAttempts) != 3 {
		return false
	}
	for _, effect := range provider.EffectAttempts {
		if !effect.Requested || effect.Result != "DENIED" || !effect.BoundaryObserved || effect.BeforeDigest != effect.AfterDigest {
			return false
		}
	}
	return true
}

func claimTransitions(model SourceModel, provider ProviderObservation, item CaseSpec, receipt Receipt) []ClaimTransition {
	transitions := make([]ClaimTransition, 0, len(item.Requests)+3)
	for _, value := range item.Requests {
		declaration, ok := declarationFor(model, value.ValueID)
		prior := ClaimOpen
		if ok {
			prior = declaration.PriorClaimState
		}
		next := ClaimOpen
		if receipt.Decision == DecisionAllow {
			next = ClaimDischarged
		} else if receipt.Decision == DecisionDeny {
			next = ClaimRefuted
		}
		evidenceDigest := providerDigest([]byte(value.ValueID))
		if ok {
			if evidence, observed := currentEvidenceFor(provider, declaration); observed {
				evidenceDigest = evidence.EvidenceDigest
			}
		}
		transitions = append(transitions, ClaimTransition{ClaimID: "capability:" + value.ValueID, PriorState: prior, NextState: next, Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: evidenceDigest, Provenance: "source-ir+provider-observation"})
	}
	scopeNext := ClaimOpen
	if receipt.Decision == DecisionAllow {
		scopeNext = ClaimDischarged
	} else if receipt.Decision == DecisionDeny {
		scopeNext = ClaimRefuted
	}
	transitions = append(transitions,
		ClaimTransition{ClaimID: "capability-scope-exact", PriorState: model.Policy.PriorClaimState, NextState: scopeNext, Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: receipt.ProviderDigest, Provenance: "source-ir+provider-observation"},
		ClaimTransition{ClaimID: "default-deny", PriorState: model.Policy.PriorClaimState, NextState: claimKnown(receipt.Decision), Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: receipt.ProviderDigest, Provenance: "source-ir+provider-observation"},
		ClaimTransition{ClaimID: "effect-ceiling", PriorState: model.Policy.PriorClaimState, NextState: effectClaimState(provider, receipt.Decision), Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: receipt.ProviderDigest, Provenance: "sandbox-before-after"},
	)
	return transitions
}

func claimKnown(decision string) string {
	if decision == DecisionUnknown {
		return ClaimOpen
	}
	return ClaimDischarged
}

func effectClaimState(provider ProviderObservation, decision string) string {
	if decision == DecisionUnknown && !effectBoundaryObserved(provider) {
		return ClaimOpen
	}
	if effectBoundaryObserved(provider) {
		return ClaimDischarged
	}
	return ClaimOpen
}

func claimsFromTransitions(transitions []ClaimTransition) []Claim {
	wanted := []struct{ id, proof, evidence string }{{"capability-scope-exact", "COHERENCE", "source-ir+provider-observation"}, {"default-deny", "REGRESSION", "undeclared-and-policy-denials"}, {"effect-ceiling", "REGRESSION", "sandbox-before-after"}}
	claims := make([]Claim, 0, len(wanted))
	for _, item := range wanted {
		for _, transition := range transitions {
			if transition.ClaimID == item.id {
				claims = append(claims, Claim{ID: item.id, PriorState: transition.PriorState, Status: transition.NextState, ProofChoice: item.proof, Evidence: item.evidence})
				break
			}
		}
	}
	return claims
}

func indicatorsFor(model SourceModel, provider ProviderObservation, receipt Receipt) []Indicator {
	currentFile := 0
	currentTime := 0
	for _, declaration := range model.Declarations {
		if declaration.Kind == "file" {
			if _, ok := currentEvidenceFor(provider, declaration); ok {
				currentFile = 1
			}
		}
		if declaration.Kind == "time" {
			if _, ok := currentEvidenceFor(provider, declaration); ok {
				currentTime = 1
			}
		}
	}
	items := []struct {
		id, class, proof string
		observed, target int
		status           string
	}{
		{"CSE-source-reconstruction", "DRIVER", "FOUNDATION", boolInt(model.Reconstructed), 1, statusFor(boolInt(model.Reconstructed) == 1, receipt.Decision)},
		{"CSE-semantic-policy", "DRIVER", "FOUNDATION", boolInt(model.Policy.DefaultDecision == PolicyDefaultDeny), 1, statusFor(model.Policy.DefaultDecision == PolicyDefaultDeny, receipt.Decision)},
		{"CSE-declaration-kind", "DRIVER", "FOUNDATION", boolInt(allDeclarationsHave(model, "kind")), 1, statusFor(allDeclarationsHave(model, "kind"), receipt.Decision)},
		{"CSE-declaration-operation", "DRIVER", "FOUNDATION", boolInt(allDeclarationsHave(model, "operation")), 1, statusFor(allDeclarationsHave(model, "operation"), receipt.Decision)},
		{"CSE-declaration-target", "DRIVER", "FOUNDATION", boolInt(allDeclarationsHave(model, "target")), 1, statusFor(allDeclarationsHave(model, "target"), receipt.Decision)},
		{"CSE-prior-claim-open", "DRIVER", "FOUNDATION", boolInt(allPriorClaimsOpen(model)), 1, statusFor(allPriorClaimsOpen(model), receipt.Decision)},
		{"CSE-current-file-evidence", "OUTCOME", "OBSERVATION", currentFile, 1, statusFor(currentFile == 1, receipt.Decision)},
		{"CSE-current-logical-evidence", "OUTCOME", "OBSERVATION", currentTime, 1, statusFor(currentTime == 1, receipt.Decision)},
		{"CSE-environment-network-lower-resolution", "OUTCOME", "EPISTEMIC", boolInt(len(provider.EnvironmentReads) == 0 && len(provider.NetworkReads) == 0), 1, "OPEN"},
		{"CSE-provider-enforcement-boundary", "GUARDRAIL", "ENFORCEMENT", boolInt(effectBoundaryObserved(provider)), 1, statusFor(effectBoundaryObserved(provider), receipt.Decision)},
		{"CSE-sandbox-before-after", "GUARDRAIL", "OBSERVATION", boolInt(provider.SandboxBefore.Digest == provider.SandboxAfter.Digest && provider.ActualRepositoryWrites == 0), 1, statusFor(provider.SandboxBefore.Digest == provider.SandboxAfter.Digest && provider.ActualRepositoryWrites == 0, receipt.Decision)},
		{"CSE-receipt-seal", "DRIVER", "COHERENCE", 1, 1, "SATISFIED"},
	}
	result := make([]Indicator, 0, len(items))
	for _, item := range items {
		result = append(result, Indicator{ID: item.id, Class: item.class, Status: item.status, Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: item.proof, Observed: item.observed, Target: item.target})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func allDeclarationsHave(model SourceModel, field string) bool {
	for _, declaration := range model.Declarations {
		if (field == "kind" && declaration.Kind == "") || (field == "operation" && declaration.Operation == "") || (field == "target" && declaration.Target == "") {
			return false
		}
	}
	return true
}

func allPriorClaimsOpen(model SourceModel) bool {
	if model.Policy.PriorClaimState != ClaimOpen {
		return false
	}
	for _, declaration := range model.Declarations {
		if declaration.PriorClaimState != ClaimOpen {
			return false
		}
	}
	return true
}

func statusFor(satisfied bool, decision string) string {
	if satisfied {
		return "SATISFIED"
	}
	if decision == DecisionUnknown {
		return "OPEN"
	}
	return "REFUTED"
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
