package phaseseparation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type derivedRecord struct {
	SourceRecord
	EvidenceDigest string
	PreviousDigest string
}

type evaluation struct {
	Decision         string
	Cases            []CaseResult
	Transitions      []ClaimTransition
	Summary          Summary
	Indicators       []Indicator
	Views            []View
	Proofs           []Proof
	TransitionDigest string
}

// Build reads all source authorities before constructing a receipt. The
// receipt is an observation of those authorities, not an input to evaluation.
func Build(sourcePath string, sourceBytes []byte, leaksPath string, leaksBytes []byte, unknownPath string, unknownBytes []byte, headSHA string, snapshot CISnapshot) Report {
	report := Report{
		Schema: Schema, Decision: DecisionUnknown, Reason: ReasonUnknownSource, Resolution: ResolutionLower,
		HeadSHA: headSHA, Toolchain: Toolchain,
		SourcePath: sourcePath, SourceDigest: digestBytes(sourceBytes),
		LeakSourcePath: leaksPath, LeakSourceDigest: digestBytes(leaksBytes),
		UnknownSourcePath: unknownPath, UnknownSourceDigest: digestBytes(unknownBytes),
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID, ProofChoice: ProofChoiceID,
		CISnapshot: snapshot, Authority: Authority{Execution: snapshot.ExecutionAuthority, Mutation: snapshot.MutationAuthority, Promotion: snapshot.PromotionAuthority},
		Coordinate: Coordinate{Stage: "SOURCE", Step: "DECODE", Reason: ReasonUnknownSource},
	}
	mainFile, err := ParseAndLower(sourcePath, sourceBytes)
	if err != nil {
		return finalize(report)
	}
	leakFile, err := ParseAndLower(leaksPath, leaksBytes)
	if err != nil {
		return finalize(report)
	}
	unknownFile, err := ParseAndLower(unknownPath, unknownBytes)
	if err != nil {
		report.Unknown = UnknownResult{Decision: DecisionUnknown, Coordinate: Coordinate{Stage: "SOURCE", Step: "PARSE", Reason: ReasonUnknownSyntax}, ClaimState: StateOpen, PreviousDigest: zeroDigest}
		return finalize(report)
	}
	if mainFile.IR.Package != leakFile.IR.Package || mainFile.IR.Namespace != leakFile.IR.Namespace || mainFile.IR.Package != unknownFile.IR.Package || mainFile.IR.Namespace != unknownFile.IR.Namespace {
		return finalize(report)
	}

	base := evaluateCorpus(mainFile.Activities, leakFile.Activities)
	report.Cases, report.Transitions, report.Summary = base.Cases, base.Transitions, base.Summary
	report.Indicators, report.Views, report.Proofs = buildEvidence(base, report)
	report.SemanticIntervention = semanticIntervention(mainFile, leakFile, base)
	report.NonsemanticIntervention = nonsemanticIntervention(mainFile, leakFile, base)
	report.Summary.IndicatorsTotal = len(report.Indicators)
	report.Summary.IndicatorsSatisfied = countSatisfied(report.Indicators)
	report.Summary.SemanticCausality = report.SemanticIntervention.Numerator
	report.Summary.SemanticCausalityTotal = report.SemanticIntervention.Denominator
	report.Summary.NonsemanticPreservation = report.NonsemanticIntervention.Numerator
	report.Summary.NonsemanticPreservationTotal = report.NonsemanticIntervention.Denominator
	report.Indicators, report.Views, report.Proofs = buildEvidence(base, report)
	report.Summary.IndicatorsTotal = len(report.Indicators)
	report.Summary.IndicatorsSatisfied = countSatisfied(report.Indicators)
	report.Unknown = deriveUnknown(unknownFile.Activities)
	report.Summary.UnknownCases = boolToInt(report.Unknown.Decision == DecisionUnknown)
	report.Summary.RepositoryWrites = snapshot.RepositoryWrites
	if exactReport(report) {
		report.Decision = DecisionPass
		report.Reason = ReasonExact
		report.Resolution = ResolutionExact
		report.Coordinate = Coordinate{Stage: "EXECUTION", Step: "ADJUDICATE", Reason: ReasonExact}
	} else {
		report.Decision = DecisionUnknown
		report.Reason = ReasonUnknownContract
		report.Coordinate = Coordinate{Stage: "EXECUTION", Step: "ADJUDICATE", Reason: ReasonUnknownContract}
	}
	return finalize(report)
}

func evaluateCorpus(mainRecords, leakRecords []SourceRecord) evaluation {
	all := append(append([]SourceRecord(nil), mainRecords...), leakRecords...)
	annotated := annotateRecords(all)
	byCase := make(map[string][]derivedRecord)
	for _, record := range annotated {
		byCase[record.CaseKey] = append(byCase[record.CaseKey], record)
	}
	result := evaluation{Decision: DecisionUnknown}
	result.Summary.SourceCasesProcessed = len(byCase)
	result.Summary.SourceCasesTotal = len(byCase)
	result.Summary.CleanCasesTotal = boolToInt(len(byCase["clean"]) > 0)
	result.Summary.LeakageRejectionsTotal = len(leakRecords)
	result.Summary.ClaimTransitionsTotal = len(byCase["clean"])
	result.Summary.ExplicitClaimTransfersTotal = len(byCase["clean"])
	cleanResult, transitions := evaluateCleanCase(byCase["clean"])
	result.Cases = append(result.Cases, cleanResult)
	result.Transitions = transitions
	result.Summary.CleanCasesPassed = boolToInt(cleanResult.Passed)
	result.Summary.ClaimTransitionsPreserved = countPreserved(transitions)
	result.Summary.ExplicitClaimTransfers = countPreserved(transitions)
	for _, name := range []string{"value-leak", "authority-leak", "evidence-leak", "phase-skip", "phase-reverse"} {
		caseResult := evaluateLeakCase(name, byCase[name])
		result.Cases = append(result.Cases, caseResult)
		result.Summary.LeakageRejections += boolToInt(caseResult.Passed)
	}
	result.TransitionDigest = digestValue(result.Transitions)
	if result.Summary.CleanCasesPassed == 1 && result.Summary.LeakageRejections == result.Summary.LeakageRejectionsTotal && result.Summary.ClaimTransitionsPreserved == result.Summary.ClaimTransitionsTotal && result.Summary.ExplicitClaimTransfers == result.Summary.ExplicitClaimTransfersTotal {
		result.Decision = DecisionPass
	}
	return result
}

func annotateRecords(records []SourceRecord) []derivedRecord {
	result := make([]derivedRecord, 0, len(records))
	previous := zeroDigest
	for _, record := range records {
		evidence := evidenceDigest(record)
		result = append(result, derivedRecord{SourceRecord: record, EvidenceDigest: evidence, PreviousDigest: previous})
		previous = evidence
	}
	return result
}

func evaluateCleanCase(records []derivedRecord) (CaseResult, []ClaimTransition) {
	result := caseResultFromRecords("clean", "CLEAN", records)
	transitions := make([]ClaimTransition, 0, len(records))
	valid := len(records) == ExpectedClaimTransitions
	for index, record := range records {
		transition, reason := deriveClaimTransition(record)
		if index == 0 && record.ClaimStateTo != StateOpen {
			valid = false
		}
		if index == len(records)-1 && record.ClaimStateTo != StateDischarged {
			valid = false
		}
		if reason != "" {
			valid = false
		}
		transitions = append(transitions, transition)
	}
	if valid {
		result.Outcome = StateDischarged
		result.ClaimState = StateDischarged
		result.Reason = "EXPLICIT_CLAIM_DISCHARGED"
		result.Passed = true
	} else {
		result.Outcome = StateRefuted
		result.ClaimState = StateRefuted
		result.Reason = "CLAIM_TRANSFER_NOT_PRESERVED"
	}
	result.Stage, result.Step = "EXECUTION", "ADJUDICATE"
	return result, transitions
}

func evaluateLeakCase(name string, records []derivedRecord) CaseResult {
	result := caseResultFromRecords(name, "LEAKAGE", records)
	if len(records) != 1 {
		result.Outcome, result.ClaimState, result.Reason = StateRefuted, StateRefuted, ReasonUnknownContract
		result.Stage, result.Step = "EXECUTION", "ADJUDICATE"
		return result
	}
	reason := deriveLeakReason(records[0])
	result.Outcome, result.ClaimState, result.Reason = StateRefuted, StateRefuted, reason
	result.Stage, result.Step = "EXECUTION", "ADJUDICATE"
	result.Passed = isKnownLeakReason(reason)
	return result
}

func caseResultFromRecords(name, class string, records []derivedRecord) CaseResult {
	result := CaseResult{Name: name, Class: class, TransferCount: len(records)}
	for _, record := range records {
		result.TransferIDs = append(result.TransferIDs, record.TransferID)
		result.ValueIDs = append(result.ValueIDs, record.ValueID, record.ToValueID)
		result.PayloadClasses = append(result.PayloadClasses, record.PayloadClass)
		result.EvidenceDigests = append(result.EvidenceDigests, record.EvidenceDigest)
		result.Provenances = append(result.Provenances, record.Provenance)
		result.PreviousDigests = append(result.PreviousDigests, record.PreviousDigest)
	}
	return result
}

func deriveClaimTransition(record derivedRecord) (ClaimTransition, string) {
	transition := ClaimTransition{
		ID: record.TransferID, FromPhase: record.FromPhase, ToPhase: record.ToPhase,
		FromClaim: record.FromValueID, ToClaim: record.ToValueID,
		FromState: record.ClaimStateFrom, ToState: record.ClaimStateTo,
		ClaimDigest: claimDigest(record.SourceRecord), TargetDigest: targetDigest(record.SourceRecord),
		EvidenceDigest: record.EvidenceDigest, Provenance: record.Provenance, PreviousDigest: record.PreviousDigest,
		MetaOperation: MetaOperationID, ProofChoice: ProofChoiceID,
	}
	if record.PayloadClass != PayloadClaim {
		return transition, leakReason(record.PayloadClass)
	}
	if !allowedClaimEdge(record.FromPhase, record.ToPhase) {
		return transition, edgeReason(record.FromPhase, record.ToPhase)
	}
	if record.ClaimDigest != transition.ClaimDigest || record.TargetDigest != transition.TargetDigest {
		return transition, "CLAIM_DIGEST_MISMATCH"
	}
	if record.Provenance == "" || record.ClaimStateFrom != StateOpen {
		return transition, "CLAIM_PROVENANCE_MISMATCH"
	}
	transition.Preserved = true
	return transition, ""
}

func deriveLeakReason(record derivedRecord) string {
	if record.PayloadClass != PayloadClaim {
		return leakReason(record.PayloadClass)
	}
	return edgeReason(record.FromPhase, record.ToPhase)
}

func allowedClaimEdge(from, to string) bool {
	return (from == "source" && to == "expansion") || (from == "expansion" && to == "execution")
}

func edgeReason(from, to string) string {
	if from == "source" && to == "execution" {
		return expectedLeakReasons["phase-skip"]
	}
	if from == "execution" && to == "expansion" {
		return expectedLeakReasons["phase-reverse"]
	}
	return "PHASE_EDGE_INVALID"
}

func leakReason(payload string) string {
	return expectedLeakReasons[payload]
}

func isKnownLeakReason(reason string) bool {
	for _, expected := range expectedLeakReasons {
		if reason == expected {
			return true
		}
	}
	return false
}

func buildEvidence(base evaluation, report Report) ([]Indicator, []View, []Proof) {
	evidenceOK := true
	for _, result := range base.Cases {
		evidenceOK = evidenceOK && len(result.EvidenceDigests) == result.TransferCount && len(result.Provenances) == result.TransferCount && len(result.PreviousDigests) == result.TransferCount
	}
	claimLifecycleOK := true
	for _, transition := range base.Transitions {
		claimLifecycleOK = claimLifecycleOK && transition.FromState == StateOpen && (transition.ToState == StateOpen || transition.ToState == StateDischarged) && transition.Preserved
	}
	values := []bool{
		report.Producer == ProducerID,
		report.Consumer == ConsumerID,
		base.Summary.SourceCasesProcessed == ExpectedSourceCases && base.Summary.SourceCasesTotal == ExpectedSourceCases,
		base.Summary.CleanCasesPassed == base.Summary.CleanCasesTotal && base.Summary.CleanCasesTotal == ExpectedCleanCases,
		base.Summary.LeakageRejections == base.Summary.LeakageRejectionsTotal && base.Summary.LeakageRejectionsTotal == ExpectedLeakageCases,
		base.Summary.ClaimTransitionsPreserved == base.Summary.ClaimTransitionsTotal && base.Summary.ClaimTransitionsTotal == ExpectedClaimTransitions,
		evidenceOK,
		claimLifecycleOK,
		report.SemanticIntervention.Passed,
		report.NonsemanticIntervention.Passed,
		report.CISnapshot.RepositoryWrites == 0 && !report.CISnapshot.MutationAuthority,
		!report.CISnapshot.PromotionAuthority && !report.CISnapshot.ExecutionAuthority,
	}
	indicators := make([]Indicator, 0, len(values))
	for index, satisfied := range values {
		indicators = append(indicators, Indicator{ID: fmt.Sprintf("PHASE-%02d", index+1), MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Numerator: boolToInt(satisfied), Denominator: 1, Satisfied: satisfied})
	}
	views := []View{
		{Audience: "PRODUCER", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: base.Summary.SourceCasesProcessed, Total: base.Summary.SourceCasesTotal},
		{Audience: "CONSUMER", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: base.Summary.CleanCasesPassed + base.Summary.LeakageRejections, Total: base.Summary.CleanCasesTotal + base.Summary.LeakageRejectionsTotal},
		{Audience: "GOVERNOR", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: countSatisfied(indicators), Total: len(indicators)},
	}
	for index := range views {
		views[index].BasisPoints = ratioBasisPoints(views[index].Satisfied, views[index].Total)
	}
	proofs := []Proof{
		{Choice: "BOUNDARY", Claim: "phase-local values reject non-claim payload authority", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(base.Cases), Provenance: report.SourcePath, Passed: base.Summary.LeakageRejections == base.Summary.LeakageRejectionsTotal},
		{Choice: "TRANSITION", Claim: "only digest-bound adjacent claims discharge", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(base.Transitions), Provenance: report.SourcePath, Passed: base.Summary.ClaimTransitionsPreserved == base.Summary.ClaimTransitionsTotal},
		{Choice: "AUTHORITY", Claim: "CI observes zero repository writes and mutation authority", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(report.CISnapshot), Provenance: report.CISnapshot.Permissions, Passed: report.CISnapshot.RepositoryWrites == 0 && !report.CISnapshot.MutationAuthority && !report.CISnapshot.PromotionAuthority},
	}
	return indicators, views, proofs
}

func deriveUnknown(parsed []SourceRecord) UnknownResult {
	if len(parsed) != 1 {
		return UnknownResult{Decision: DecisionUnknown, Coordinate: Coordinate{Stage: "SOURCE", Step: "DECODE", Reason: ReasonUnknownContract}, ClaimState: StateOpen, PreviousDigest: zeroDigest}
	}
	record := parsed[0]
	return UnknownResult{Decision: DecisionUnknown, Coordinate: Coordinate{Stage: record.Stage, Step: record.Step, Reason: record.DeclaredReason}, ClaimState: StateOpen, EvidenceDigest: evidenceDigest(record), Provenance: record.Provenance, PreviousDigest: zeroDigest}
}

func semanticIntervention(mainFile, leakFile ParsedFile, base evaluation) Intervention {
	result := Intervention{Kind: "SEMANTIC", Denominator: 1}
	variantSource := bytes.Replace(mainFile.Source, []byte("payload_class=claim"), []byte("payload_class=value"), 1)
	variantSource = bytes.Replace(variantSource, []byte("claim_digest=sha256:ee3e8ebaa490d076fa230325fe30ef061c629c7cd2e5d5ef41df5e52a06548c3"), []byte("claim_digest=none"), 1)
	variantSource = bytes.Replace(variantSource, []byte("target_digest=sha256:f71af2266668baa89342e7740fe22b77b58f5aa612fe716b7fe9257380fb34fa"), []byte("target_digest=none"), 1)
	variant, err := ParseAndLower(mainFile.Filename, variantSource)
	if err != nil {
		return result
	}
	variantEvaluation := evaluateCorpus(variant.Activities, leakFile.Activities)
	result.BaseSemanticDigest, result.VariantSemanticDigest = mainFile.SemanticHash, variant.SemanticHash
	result.BaseDecision, result.VariantDecision = base.Decision, variantEvaluation.Decision
	result.BaseTransitionDigest, result.VariantTransitionDigest = base.TransitionDigest, variantEvaluation.TransitionDigest
	result.Changed = result.BaseSemanticDigest != result.VariantSemanticDigest && (result.BaseDecision != result.VariantDecision || result.BaseTransitionDigest != result.VariantTransitionDigest)
	result.Numerator = boolToInt(result.Changed)
	result.Passed = result.Changed
	return result
}

func nonsemanticIntervention(mainFile, leakFile ParsedFile, base evaluation) Intervention {
	result := Intervention{Kind: "NONSEMANTIC", Denominator: 1}
	variantSource := append(append([]byte(nil), mainFile.Source...), []byte("\n// nonsemantic intervention\n")...)
	variant, err := ParseAndLower(mainFile.Filename, variantSource)
	if err != nil {
		return result
	}
	variantEvaluation := evaluateCorpus(variant.Activities, leakFile.Activities)
	result.BaseSemanticDigest, result.VariantSemanticDigest = mainFile.SemanticHash, variant.SemanticHash
	result.BaseDecision, result.VariantDecision = base.Decision, variantEvaluation.Decision
	result.BaseTransitionDigest, result.VariantTransitionDigest = base.TransitionDigest, variantEvaluation.TransitionDigest
	result.Preserved = result.BaseSemanticDigest == result.VariantSemanticDigest && result.BaseDecision == result.VariantDecision && result.BaseTransitionDigest == result.VariantTransitionDigest
	result.Numerator = boolToInt(result.Preserved)
	result.Passed = result.Preserved
	return result
}

func exactReport(report Report) bool {
	return report.Decision == DecisionUnknown &&
		report.Summary.SourceCasesProcessed == ExpectedSourceCases &&
		report.Summary.SourceCasesTotal == ExpectedSourceCases &&
		report.Summary.CleanCasesPassed == ExpectedCleanCases &&
		report.Summary.CleanCasesTotal == ExpectedCleanCases &&
		report.Summary.LeakageRejections == ExpectedLeakageCases &&
		report.Summary.LeakageRejectionsTotal == ExpectedLeakageCases &&
		report.Summary.ClaimTransitionsPreserved == ExpectedClaimTransitions &&
		report.Summary.ClaimTransitionsTotal == ExpectedClaimTransitions &&
		report.Summary.ExplicitClaimTransfers == ExpectedClaimTransitions &&
		report.Summary.ExplicitClaimTransfersTotal == ExpectedClaimTransitions &&
		report.Summary.IndicatorsSatisfied == ExpectedIndicators &&
		report.Summary.IndicatorsTotal == ExpectedIndicators &&
		report.Summary.SemanticCausality == 1 && report.Summary.SemanticCausalityTotal == 1 &&
		report.Summary.NonsemanticPreservation == 1 && report.Summary.NonsemanticPreservationTotal == 1 &&
		report.Summary.UnknownCases == 1 && report.Summary.RepositoryWrites == 0 &&
		report.Authority == (Authority{}) && report.Unknown.Decision == DecisionUnknown && report.Unknown.ClaimState == StateOpen &&
		len(report.Cases) == ExpectedCleanCases+ExpectedLeakageCases && len(report.Transitions) == ExpectedClaimTransitions && len(report.Indicators) == ExpectedIndicators && len(report.Views) == ExpectedViews && len(report.Proofs) == ExpectedProofs &&
		allPassed(report.Indicators) && allPassed(report.Proofs)
}

func allPassed(values interface{}) bool {
	switch typed := values.(type) {
	case []Indicator:
		for _, value := range typed {
			if !value.Satisfied || value.Numerator != value.Denominator {
				return false
			}
		}
	case []Proof:
		for _, value := range typed {
			if !value.Passed {
				return false
			}
		}
	}
	return true
}

func countPreserved(transitions []ClaimTransition) int {
	count := 0
	for _, transition := range transitions {
		count += boolToInt(transition.Preserved)
	}
	return count
}

func countSatisfied(indicators []Indicator) int {
	count := 0
	for _, indicator := range indicators {
		count += boolToInt(indicator.Satisfied)
	}
	return count
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func ratioBasisPoints(numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return numerator * 10000 / denominator
}

func claimDigest(record SourceRecord) string {
	canonical := strings.Join([]string{"claim", record.CaseKey, record.TransferID, record.FromValueID, record.ToValueID, record.FromPhase, record.ToPhase, record.FromLiteralClass, record.ToLiteralClass, record.Provenance}, "|")
	return digestString(canonical)
}

func targetDigest(record SourceRecord) string {
	return digestString(strings.Join([]string{"target", record.ToValueID, record.ToLiteralClass}, "|"))
}

func evidenceDigest(record SourceRecord) string {
	return digestString("source-value-program|" + record.Program)
}

func digestBytes(value []byte) string {
	return digestString(string(value))
}

func digestValue(value interface{}) string {
	encoded, _ := json.Marshal(value)
	return digestString(string(encoded))
}

func digestString(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func finalize(report Report) Report {
	report.Digest = ""
	encoded, _ := json.Marshal(report)
	report.Digest = digestString(string(encoded))
	return report
}
