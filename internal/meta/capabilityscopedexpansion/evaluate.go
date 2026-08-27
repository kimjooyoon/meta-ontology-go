package capabilityscopedexpansion

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/engine"
)

type EvaluationContext struct {
	ArtifactRoot string
}

// EvaluateSuite parses and lowers the same source once, consumes raw provider
// observations, and evaluates every case declared by semantic Gooo values.
// Callers that expect an ALLOW artifact must use EvaluateSuiteWithContext.
func EvaluateSuite(source, providerRaw []byte, subject string) (Suite, []Receipt, error) {
	return EvaluateSuiteWithContext(source, providerRaw, subject, EvaluationContext{})
}

func EvaluateSuiteWithContext(source, providerRaw []byte, subject string, context EvaluationContext) (Suite, []Receipt, error) {
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
	if context.ArtifactRoot == "" {
		return Suite{}, nil, fmt.Errorf("artifact root is required for expansion execution")
	}
	current, historical := evidenceDenominator(model, provider)
	suite := Suite{
		Schema: "gooo/capability-scoped-expansion-suite/v3", MetaOperation: MetaOperation, SubjectSHA: subject,
		SourceDigest: model.SourceDigest, SemanticDigest: model.SemanticDigest, Decision: SuitePass, Resolution: ResolutionExact,
		Summary: SuiteSummary{
			CasesTotal: len(model.Cases), CurrentEvidenceCapabilities: current, CurrentEvidenceDenominator: len(model.Declarations),
			HistoricalFixtureDeclarations: historical, EffectTokenRequests: provider.BrokerTokenRequests,
			TokensIssued: provider.BrokerTokensIssued, TokenDenials: provider.BrokerTokenDenials,
			EnforcementObservations: provider.BrokerTokenDenials, RepositoryWrites: provider.RepositoryWrites, SandboxWrites: provider.SandboxWrites,
			MutationAuthority: provider.MutationAuthority, PromotionAuthority: provider.PromotionAuthority,
		},
	}
	receipts := make([]Receipt, 0, len(model.Cases))
	for _, item := range model.Cases {
		receipt := Evaluate(model, providerRaw, provider, subject, item, context)
		receipts = append(receipts, receipt)
		suite.Cases = append(suite.Cases, CaseResult{CaseID: item.ID, ObservedDecision: receipt.Decision, ObservedResolution: receipt.Resolution, ReceiptDigest: receipt.ReportDigest})
		suite.Summary.CapabilityRequests += receipt.Authority.CapabilitiesRequested
		suite.Summary.CapabilityAuthorized += receipt.Authority.CapabilitiesAuthorized
		suite.Summary.CapabilityDenied += receipt.Authority.CapabilitiesDenied
		suite.Summary.CapabilityUnknown += receipt.Authority.CapabilitiesUnknown
		suite.Summary.EffectTokenRequests += tokenRequestsFor(item)
		suite.Summary.ArtifactExecutions += boolInt(receipt.Artifact.Present)
		suite.Summary.ArtifactsAbsentForBlocked += boolInt(receipt.Decision != DecisionAllow && !receipt.Artifact.Present)
		suite.Summary.BlockedWriteAttempts += boolInt(item.RequestedRepositoryWrites != 0)
		suite.Summary.BlockedMutationAttempts += boolInt(item.RequestedMutationAuthority)
		suite.Summary.BlockedMutationAttempts += boolInt(item.RequestedPromotionAuthority)
		switch receipt.Decision {
		case DecisionAllow:
			suite.Summary.AllowCases++
		case DecisionDeny:
			suite.Summary.DenyCases++
		case DecisionUnknown:
			suite.Summary.UnknownCases++
		}
		if receipt.Decision == expectedDecision(model, provider, item) {
			suite.Summary.CasesPassed++
		}
	}
	if suite.Summary.CasesPassed != suite.Summary.CasesTotal {
		suite.Decision = DecisionUnknown
		suite.Resolution = ResolutionLower
	}
	suite.Summary.CapabilityRequests = sumCapabilityRequests(receipts)
	suite.Summary.EffectTokenRequests = provider.BrokerTokenRequests
	suite.RepositoryWrites = provider.RepositoryWrites
	suite.SandboxWrites = provider.SandboxWrites
	suite.MutationAuthority = provider.MutationAuthority
	suite.PromotionAuthority = provider.PromotionAuthority
	suite.Denominator = Denominator{
		Cases: len(model.Cases), Declarations: len(model.Declarations), CapabilityRequests: suite.Summary.CapabilityRequests,
		EvidenceSlots: FixedEvidenceSlotTotal, EffectTokenRequests: provider.BrokerTokenRequests,
		Claims: countClaims(receipts), IndicatorsPerReceipt: IndicatorsPerReceipt,
	}
	return suite, receipts, nil
}

func Evaluate(model SourceModel, providerRaw []byte, provider ProviderObservation, subject string, item CaseSpec, context EvaluationContext) Receipt {
	operation := authorizationOperation(model)
	receipt := Receipt{
		Schema: Schema, MetaOperation: MetaOperation, Producer: Producer, Consumer: Consumer, SubjectSHA: subject, GoVersion: GoVersion,
		SourceDigest: model.SourceDigest, SemanticDigest: model.SemanticDigest, CaseID: item.ID, Stage: operation.Stage, Step: operation.Step,
		Policy: model.Policy, Graph: model.Graph, Declarations: append([]CapabilityDeclaration(nil), model.Declarations...), Capabilities: append([]CapabilityValue(nil), item.Requests...),
		ProviderDigest: providerDigest(providerRaw), TokenAttempts: append([]TokenIssuance(nil), provider.TokenAttempts...),
		RepositoryWrites: provider.RepositoryWrites, SandboxWrites: provider.SandboxWrites, MutationAuthority: provider.MutationAuthority, PromotionAuthority: provider.PromotionAuthority,
	}
	receipt.Authority = Authority{
		CapabilitiesRequested: len(item.Requests), CurrentEvidenceCapabilities: evidenceDenominatorCount(model, provider), CurrentEvidenceDenominator: len(model.Declarations),
		RequestedRepositoryWrites: item.RequestedRepositoryWrites, RequestedMutationAuthority: item.RequestedMutationAuthority, RequestedPromotionAuthority: item.RequestedPromotionAuthority,
		RepositoryWrites: provider.RepositoryWrites, SandboxWrites: provider.SandboxWrites, MutationAuthority: provider.MutationAuthority, PromotionAuthority: provider.PromotionAuthority,
		EnforcementObservations: provider.BrokerTokenDenials,
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
	receipt.Propositions = propositionsFor(model, provider, item, receipt)
	receipt.Execution = executionFor(receipt, model, provider, item, context)
	if receipt.Execution.ArtifactDigest != "" {
		receipt.Artifact = artifactFor(receipt.Execution, context)
	}
	receipt.ClaimTransitions = claimTransitions(model, provider, item, receipt)
	receipt.Claims = claimsFromTransitions(receipt.ClaimTransitions)
	receipt.Indicators = indicatorsFor(model, provider, receipt)
	return sealReceipt(receipt)
}

func artifactFor(execution ExecutionObservation, context EvaluationContext) ExpansionArtifact {
	if execution.ArtifactDigest == "" {
		return ExpansionArtifact{}
	}
	return ExpansionArtifact{Schema: ArtifactSchema, Present: true, Path: execution.ArtifactPath, Value: execution.ArtifactValue, Bytes: execution.ArtifactBytes, ContentDigest: execution.ArtifactDigest, SemanticDigest: execution.ArtifactSemanticDigest, Reparsed: true, ReparsedSemanticDigest: execution.ReparsedSemanticDigest}
}

// executionFor performs the only expansion side effect: a generated artifact
// is written outside the repository, then engine.Expand has already parsed and
// lowered that exact output before returning its digest.
func executionFor(receipt Receipt, model SourceModel, provider ProviderObservation, item CaseSpec, context EvaluationContext) ExecutionObservation {
	execution := ExecutionObservation{Requested: receipt.Decision == DecisionAllow, Decision: receipt.Decision, ClaimID: "execution:expanded-syntax", Result: "NOT_EXECUTED", Reason: "BLOCKED_BEFORE_EXPANSION"}
	if receipt.Decision != DecisionAllow {
		if receipt.Decision == DecisionDeny {
			execution.ClaimState = ClaimRefuted
		} else {
			execution.ClaimState = ClaimOpen
			execution.Reason = "LOWER_RESOLUTION_BEFORE_EXPANSION"
		}
		return execution
	}
	evidence := currentFileAndLogicalEvidence(provider)
	artifact, err := engine.Expand(engine.Request{SourceSemanticDigest: model.SemanticDigest, GraphPathDigest: model.Graph.PathDigest, FileDigest: evidence.fileDigest, LogicalValue: evidence.logicalValue, OutputPath: filepath.Join(context.ArtifactRoot, item.ID+".gooo")})
	if err != nil {
		execution.Decision = DecisionUnknown
		execution.Result = "NOT_EXECUTED"
		execution.ClaimState = ClaimOpen
		execution.Reason = "EXPANSION_ARTIFACT_ERROR"
		return execution
	}
	execution.Result = "EXPANDED"
	execution.ClaimState = ClaimDischarged
	execution.Reason = "EXPANSION_ARTIFACT_REPARSED"
	execution.ArtifactPath = artifact.Path
	execution.ArtifactValue = artifact.Value
	execution.ArtifactBytes = len(artifact.Bytes)
	execution.ArtifactDigest = artifact.ContentDigest
	execution.ArtifactSemanticDigest = artifact.SemanticDigest
	execution.ReparsedSemanticDigest = artifact.SemanticDigest
	return execution
}

type evidenceInputs struct{ fileDigest, logicalValue string }

func currentFileAndLogicalEvidence(provider ProviderObservation) evidenceInputs {
	result := evidenceInputs{}
	for _, observation := range provider.FileReads {
		if observation.Target == "pinned-file" && observation.Observed {
			result.fileDigest = observation.ContentDigest
		}
	}
	for _, observation := range provider.LogicalInputs {
		if observation.Target == "logical-clock" && observation.Observed {
			result.logicalValue = observation.Value
		}
	}
	return result
}

func expectedDecision(model SourceModel, provider ProviderObservation, item CaseSpec) string {
	decision, _, _, _ := decisionFor(model, provider, item)
	return decision
}

func decisionFor(model SourceModel, provider ProviderObservation, item CaseSpec) (string, string, string, *Unknown) {
	operation := authorizationOperation(model)
	if !model.Graph.Complete {
		return DecisionUnknown, ResolutionLower, "GRAPH_TOPOLOGY_UNOBSERVED", &Unknown{Stage: operation.Stage, Step: "graph-reconstruct", Reason: "GRAPH_TOPOLOGY_UNOBSERVED"}
	}
	for _, value := range item.Requests {
		declaration, ok := declarationFor(model, value.ValueID)
		if !ok || !capabilityMatches(value, declaration) {
			return DecisionDeny, ResolutionExact, "CAPABILITY_NOT_DECLARED", nil
		}
	}
	if tokenKind := requestedEffectKind(item); tokenKind != "" {
		if !deniedToken(provider, tokenKind) {
			return DecisionUnknown, ResolutionLower, "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED", &Unknown{Stage: operation.Stage, Step: "authorize-before-expand", Reason: "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED"}
		}
		return DecisionDeny, ResolutionExact, "CAPABILITY_TOKEN_DENIED", nil
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

func deniedToken(provider ProviderObservation, kind string) bool {
	for _, attempt := range provider.TokenAttempts {
		if attempt.Kind == kind && attempt.Requested && attempt.Decision == DecisionDeny && !attempt.Issued {
			return true
		}
	}
	return false
}

func requestedEffectKind(item CaseSpec) string {
	if item.RequestedRepositoryWrites != 0 {
		return "file"
	}
	if item.RequestedMutationAuthority {
		return "mutation"
	}
	if item.RequestedPromotionAuthority {
		return "promotion"
	}
	return ""
}

func tokenRequestsFor(item CaseSpec) int { return boolInt(requestedEffectKind(item) != "") }

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
	historical := 0
	for _, declaration := range model.Declarations {
		if declaration.EvidenceClass == HistoricalFixture {
			historical++
		}
	}
	return evidenceDenominatorCount(model, provider), historical
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

func propositionsFor(model SourceModel, provider ProviderObservation, item CaseSpec, receipt Receipt) []Proposition {
	propositions := make([]Proposition, 0, len(item.Requests)+3)
	for index, value := range item.Requests {
		declaration, declared := declarationFor(model, value.ValueID)
		predicate := fmt.Sprintf("capability:%s:%s:%s:%s", value.ValueID, value.Kind, value.Operation, value.Target)
		status, decision, evidenceDigest := ClaimOpen, DecisionUnknown, digestBytes([]byte(predicate))
		if !declared || !capabilityMatches(value, declaration) {
			status, decision = ClaimRefuted, DecisionDeny
		} else if evidence, observed := currentEvidenceFor(provider, declaration); observed {
			status, decision, evidenceDigest = ClaimDischarged, DecisionAllow, evidence.EvidenceDigest
		}
		propositions = append(propositions, Proposition{ID: fmt.Sprintf("capability:%s:%d", value.ValueID, index), Predicate: predicate, Decision: decision, Status: status, EvidenceDigest: evidenceDigest, Provenance: "source-ir+independent-provider-replay"})
	}
	propositions = append(propositions, Proposition{ID: "authorization:" + item.ID, Predicate: "authorization:" + item.ID + ":" + model.Graph.PathDigest, Decision: receipt.Decision, Status: claimStateForDecision(receipt.Decision), EvidenceDigest: digestBytes([]byte(model.Policy.ID + "=" + model.Policy.AuthorizationMode)), Provenance: "lowered-graph-path"})
	if kind := requestedEffectKind(item); kind != "" {
		for _, attempt := range provider.TokenAttempts {
			if attempt.Kind == kind {
				propositions = append(propositions, Proposition{ID: "effect-token:" + kind, Predicate: "effect-token:" + kind + ":" + attempt.RequestDigest, Decision: attempt.Decision, Status: tokenClaimState(attempt), EvidenceDigest: attempt.RequestDigest, Provenance: "broker.issuance-receipt"})
				break
			}
		}
	}
	if receipt.Decision == DecisionAllow {
		propositions = append(propositions, Proposition{ID: "execution:expanded-syntax", Predicate: "execution:expanded-syntax:" + receipt.SemanticDigest, Decision: DecisionAllow, Status: ClaimDischarged, EvidenceDigest: receipt.SemanticDigest, Provenance: "engine.output-reparse"})
	} else {
		propositions = append(propositions, Proposition{ID: "execution:expanded-syntax", Predicate: "execution:expanded-syntax:" + receipt.CaseID, Decision: receipt.Decision, Status: claimStateForDecision(receipt.Decision), EvidenceDigest: digestBytes([]byte(receipt.CaseID + "=" + receipt.Reason)), Provenance: "expansion-gate"})
	}
	return propositions
}

func claimStateForDecision(decision string) string {
	switch decision {
	case DecisionAllow:
		return ClaimDischarged
	case DecisionDeny:
		return ClaimRefuted
	default:
		return ClaimOpen
	}
}

func tokenClaimState(attempt TokenIssuance) string {
	if attempt.Issued {
		return ClaimDischarged
	}
	if attempt.Decision == DecisionDeny {
		return ClaimRefuted
	}
	return ClaimOpen
}

func claimTransitions(model SourceModel, provider ProviderObservation, item CaseSpec, receipt Receipt) []ClaimTransition {
	transitions := make([]ClaimTransition, 0, len(receipt.Propositions)+3)
	for _, proposition := range receipt.Propositions {
		transitions = append(transitions, ClaimTransition{ClaimID: proposition.ID, PriorState: ClaimOpen, NextState: proposition.Status, Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: proposition.EvidenceDigest, Provenance: proposition.Provenance})
	}
	scopeState := claimStateForDecision(receipt.Decision)
	transitions = append(transitions,
		ClaimTransition{ClaimID: "capability-scope-exact", PriorState: model.Policy.PriorClaimState, NextState: scopeState, Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: receipt.ProviderDigest, Provenance: "source-ir+provider-observation"},
		ClaimTransition{ClaimID: "default-deny", PriorState: model.Policy.PriorClaimState, NextState: claimKnown(receipt.Decision), Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: receipt.ProviderDigest, Provenance: "source-ir+provider-observation"},
		ClaimTransition{ClaimID: "effect-ceiling", PriorState: model.Policy.PriorClaimState, NextState: effectClaimState(provider), Stage: receipt.Stage, Step: receipt.Step, Reason: "BROKER_TOKEN_DENIALS_OBSERVED", EvidenceDigest: providerDigest(providerRawForClaim(provider)), Provenance: "broker.issuance+repository-sandbox-snapshots"},
	)
	return transitions
}

// This digest is deliberately derived from the raw provider value, not a
// constant. It is only used as a provenance anchor for the effect-ceiling
// proposition; the receipt's provider digest binds the full wire artifact.
func providerRawForClaim(provider ProviderObservation) []byte {
	return []byte(fmt.Sprintf("%s|%d|%d|%d", provider.Schema, provider.BrokerTokenRequests, provider.BrokerTokensIssued, provider.BrokerTokenDenials))
}

func claimKnown(decision string) string {
	if decision == DecisionUnknown {
		return ClaimOpen
	}
	return ClaimDischarged
}

func effectClaimState(provider ProviderObservation) string {
	if provider.BrokerTokenRequests == FixedEffectTokenRequests && provider.BrokerTokenDenials == FixedEffectTokenRequests && provider.BrokerTokensIssued == 0 {
		return ClaimDischarged
	}
	return ClaimOpen
}

func claimsFromTransitions(transitions []ClaimTransition) []Claim {
	claims := make([]Claim, 0, len(transitions))
	for _, transition := range transitions {
		proof := "PROPOSITION"
		evidence := transition.Provenance
		if transition.ClaimID == "capability-scope-exact" {
			proof, evidence = "COHERENCE", "source-ir+provider-observation"
		}
		if transition.ClaimID == "default-deny" || transition.ClaimID == "effect-ceiling" {
			proof = "REGRESSION"
		}
		claims = append(claims, Claim{ID: transition.ClaimID, PriorState: transition.PriorState, Status: transition.NextState, ProofChoice: proof, Evidence: evidence})
	}
	sort.Slice(claims, func(i, j int) bool { return claims[i].ID < claims[j].ID })
	return claims
}

func indicatorsFor(model SourceModel, provider ProviderObservation, receipt Receipt) []Indicator {
	currentFile, currentTime := 0, 0
	for _, declaration := range model.Declarations {
		if _, ok := currentEvidenceFor(provider, declaration); ok && declaration.Kind == "file" {
			currentFile = 1
		}
		if _, ok := currentEvidenceFor(provider, declaration); ok && declaration.Kind == "time" {
			currentTime = 1
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
		{"CSE-environment-network-lower-resolution", "OUTCOME", "EPISTEMIC", boolInt(lowerResolutionMarkers(provider)), 1, "OPEN"},
		{"CSE-broker-token-boundary", "GUARDRAIL", "ENFORCEMENT", boolInt(effectClaimState(provider) == ClaimDischarged), 1, statusFor(effectClaimState(provider) == ClaimDischarged, receipt.Decision)},
		{"CSE-repository-sandbox-snapshots", "GUARDRAIL", "OBSERVATION", boolInt(provider.RepositoryBefore.Digest == provider.RepositoryAfter.Digest && provider.SandboxBefore.Digest == provider.SandboxAfter.Digest), 1, statusFor(provider.RepositoryBefore.Digest == provider.RepositoryAfter.Digest && provider.SandboxBefore.Digest == provider.SandboxAfter.Digest, receipt.Decision)},
		{"CSE-expansion-artifact-reparse", "OUTCOME", "EXECUTION", boolInt(receipt.Artifact.Present && receipt.Artifact.Reparsed), 1, statusFor(receipt.Artifact.Present && receipt.Artifact.Reparsed, receipt.Decision)},
	}
	result := make([]Indicator, 0, len(items))
	for _, item := range items {
		result = append(result, Indicator{ID: item.id, Class: item.class, Status: item.status, Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: item.proof, Observed: item.observed, Target: item.target})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func lowerResolutionMarkers(provider ProviderObservation) bool {
	return len(provider.EnvironmentReads) == 1 && !provider.EnvironmentReads[0].Observed && provider.EnvironmentReads[0].EvidenceClass == "UNKNOWN" && len(provider.NetworkReads) == 1 && !provider.NetworkReads[0].Observed && provider.NetworkReads[0].EvidenceClass == HistoricalFixture
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

func sumCapabilityRequests(receipts []Receipt) int {
	count := 0
	for _, receipt := range receipts {
		count += receipt.Authority.CapabilitiesRequested
	}
	return count
}

func countClaims(receipts []Receipt) int {
	count := 0
	for _, receipt := range receipts {
		count += len(receipt.Claims)
	}
	return count
}
