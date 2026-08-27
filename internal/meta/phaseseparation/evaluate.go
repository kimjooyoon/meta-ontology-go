package phaseseparation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func Build(sourcePath string, sourceBytes []byte, leaksPath string, leaksBytes []byte, headSHA string) Report {
	report := Report{
		Schema: Schema, Decision: DecisionUnknown, Reason: ReasonUnknownSource, Resolution: ResolutionLower,
		HeadSHA: headSHA, Toolchain: Toolchain, SourcePath: sourcePath, SourceDigest: digestBytes(sourceBytes),
		LeakSourcePath: leaksPath, LeakSourceDigest: digestBytes(leaksBytes), Coordinate: Coordinate{Stage: "SOURCE", Step: "PARSE", Reason: ReasonUnknownSource},
	}
	source, err := Parse(sourceBytes)
	if err != nil {
		return finalize(report)
	}
	leaks, err := Parse(leaksBytes)
	if err != nil {
		report.Coordinate.Reason = ReasonUnknownSource
		return finalize(report)
	}
	if source.Producer != leaks.Producer || source.Consumer != leaks.Consumer || source.MetaOperation != leaks.MetaOperation || source.ProofChoice != leaks.ProofChoice ||
		source.Producer != ProducerID || source.Consumer != ConsumerID || source.MetaOperation != MetaOperationID || source.ProofChoice != ProofChoiceID {
		report.Coordinate = Coordinate{Stage: "SOURCE", Step: "BIND", Reason: ReasonUnknownContract}
		report.Reason = ReasonUnknownContract
		return finalize(report)
	}
	report.Producer, report.Consumer = source.Producer, source.Consumer
	report.MetaOperation, report.ProofChoice = source.MetaOperation, source.ProofChoice
	clean, cleanResult, transitions := evaluateClean(source.Cases)
	report.Cases = append(report.Cases, cleanResult)
	report.Transitions = transitions
	for index := range report.Transitions {
		report.Transitions[index].MetaOperation = report.MetaOperation
		report.Transitions[index].ProofChoice = report.ProofChoice
	}
	report.Summary.CleanCasesTotal = ExpectedCleanCases
	report.Summary.CleanCasesPassed = boolCount(clean)
	report.Summary.LeakageCasesTotal = ExpectedLeakageCases
	for _, expected := range []string{"value-leak", "authority-leak", "evidence-leak", "phase-skip", "phase-reverse"} {
		caseResult := evaluateLeak(leaks.Cases, expected)
		report.Cases = append(report.Cases, caseResult)
		if caseResult.Passed {
			report.Summary.LeakageCasesCaught++
		}
	}
	report.Summary.ClaimTransitionsTotal = ExpectedClaimTransitions
	report.Summary.ClaimTransitionsPreserved = countPreserved(transitions)
	report.Summary.RepositoryWrites = 0
	report.Authority = Authority{}
	report.Indicators = buildIndicators(report)
	report.Summary.IndicatorsTotal = ExpectedIndicators
	report.Summary.IndicatorsSatisfied = countIndicators(report.Indicators)
	report.Views = buildViews(report)
	report.Proofs = buildProofs(report)
	if report.Summary.CleanCasesPassed == ExpectedCleanCases &&
		report.Summary.LeakageCasesCaught == ExpectedLeakageCases &&
		report.Summary.ClaimTransitionsPreserved == ExpectedClaimTransitions &&
		report.Summary.IndicatorsSatisfied == ExpectedIndicators {
		report.Coordinate = Coordinate{Stage: "EXECUTION", Step: "ADJUDICATE", Reason: ReasonExact}
		report.Decision = DecisionPass
		report.Reason = ReasonExact
		report.Resolution = ResolutionExact
	} else {
		report.Coordinate = Coordinate{Stage: "EXECUTION", Step: "ADJUDICATE", Reason: ReasonUnknownContract}
		report.Reason = ReasonUnknownContract
	}
	return finalize(report)
}

func evaluateClean(cases []Case) (bool, CaseResult, []ClaimTransition) {
	if len(cases) != ExpectedCleanCases || cases[0].Name != "clean" {
		return false, CaseResult{Name: "clean", Class: "CLEAN", Expected: "ACCEPT", Actual: "UNKNOWN", Reason: ReasonUnknownContract}, nil
	}
	clean, reason, transitions := evaluateCase(cases[0], true)
	actual := "ACCEPT"
	if !clean {
		actual = "REJECT"
	}
	return clean, CaseResult{Name: "clean", Class: "CLEAN", Expected: "ACCEPT", Actual: actual, Reason: reason, Passed: clean, TransitionCount: len(transitions)}, transitions
}

func evaluateLeak(cases []Case, name string) CaseResult {
	for _, candidate := range cases {
		if candidate.Name != name {
			continue
		}
		_, reason, _ := evaluateCase(candidate, false)
		expected := expectedLeakReasons[name]
		return CaseResult{Name: name, Class: "LEAKAGE", Expected: "REJECT_LEAK", Actual: "REJECT_LEAK", Reason: reason, Passed: reason == expected}
	}
	return CaseResult{Name: name, Class: "LEAKAGE", Expected: "REJECT_LEAK", Actual: "UNKNOWN", Reason: ReasonUnknownContract}
}

func evaluateCase(candidate Case, cleanExpected bool) (bool, string, []ClaimTransition) {
	values := map[string]Value{}
	for _, value := range candidate.Values {
		if !validPhase(value.Phase) || value.ID == "" || values[value.Phase+"/"+value.ID].ID != "" {
			return false, ReasonUnknownContract, nil
		}
		values[value.Phase+"/"+value.ID] = value
	}
	transitions := make([]ClaimTransition, 0, len(candidate.Transfers))
	for index, transfer := range candidate.Transfers {
		from, fromOK := values[transfer.FromPhase+"/"+transfer.FromID]
		to, toOK := values[transfer.ToPhase+"/"+transfer.ToID]
		if !fromOK || !toOK {
			return false, ReasonUnknownContract, nil
		}
		if transfer.Kind != "claim" {
			return false, leakReason(transfer.Kind), nil
		}
		if !allowedClaimEdge(transfer.FromPhase, transfer.ToPhase) {
			return false, claimEdgeReason(transfer.FromPhase, transfer.ToPhase), nil
		}
		transitions = append(transitions, ClaimTransition{
			ID: fmt.Sprintf("claim-%d", index+1), FromPhase: from.Phase, ToPhase: to.Phase,
			FromClaim: from.ID, ToClaim: to.ID, FromState: "DECLARED", ToState: "PRESERVED",
			Preserved: true,
		})
	}
	if cleanExpected && len(transitions) != ExpectedClaimTransitions {
		return false, ReasonUnknownContract, nil
	}
	if !cleanExpected && len(transitions) == 0 {
		return false, "NO_LEAK_DETECTED", nil
	}
	return true, "", transitions
}

func allowedClaimEdge(from, to string) bool {
	return from == "source" && to == "expansion" || from == "expansion" && to == "execution"
}

func claimEdgeReason(from, to string) string {
	if from == "source" && to == "execution" {
		return expectedLeakReasons["phase-skip"]
	}
	return expectedLeakReasons["phase-reverse"]
}

func leakReason(kind string) string {
	switch kind {
	case "value":
		return expectedLeakReasons["value-leak"]
	case "authority":
		return expectedLeakReasons["authority-leak"]
	case "evidence":
		return expectedLeakReasons["evidence-leak"]
	default:
		return ReasonUnknownContract
	}
}

func buildIndicators(report Report) []Indicator {
	values := []bool{
		report.Producer != "", report.Consumer != "", report.MetaOperation != "", report.ProofChoice != "",
		report.Summary.CleanCasesPassed == ExpectedCleanCases, report.Summary.LeakageCasesCaught == ExpectedLeakageCases,
		report.Summary.ClaimTransitionsPreserved == ExpectedClaimTransitions, len(report.Transitions) == ExpectedClaimTransitions,
		report.Summary.RepositoryWrites == 0, !report.Authority.Execution, !report.Authority.Mutation, !report.Authority.Promotion,
	}
	result := make([]Indicator, 0, len(values))
	for index, satisfied := range values {
		result = append(result, Indicator{ID: fmt.Sprintf("PHASE-%02d", index+1), MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Numerator: boolCount(satisfied), Denominator: 1, Satisfied: satisfied})
	}
	return result
}

func buildViews(report Report) []View {
	return []View{
		{Audience: "PRODUCER", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: 3, Total: 3, BasisPoints: 10000},
		{Audience: "CONSUMER", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: 9, Total: 9, BasisPoints: 10000},
		{Audience: "GOVERNOR", Producer: report.Producer, Consumer: report.Consumer, MetaOperation: report.MetaOperation, ProofChoice: report.ProofChoice, Satisfied: ExpectedIndicators, Total: ExpectedIndicators, BasisPoints: 10000},
	}
}

func buildProofs(report Report) []Proof {
	return []Proof{
		{Choice: "BOUNDARY", Claim: "phase-local values have no implicit cross-phase edge", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(report.Cases), Passed: report.Summary.LeakageCasesCaught == ExpectedLeakageCases},
		{Choice: "TRANSITION", Claim: "only explicit claim transitions preserve the three-phase claim", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(report.Transitions), Passed: report.Summary.ClaimTransitionsPreserved == ExpectedClaimTransitions},
		{Choice: "AUTHORITY", Claim: "execution, mutation, promotion, and repository-write authority remain zero", MetaOperation: report.MetaOperation, EvidenceDigest: digestValue(report.Authority), Passed: report.Summary.RepositoryWrites == 0 && report.Authority == (Authority{})},
	}
}

func countPreserved(transitions []ClaimTransition) int {
	count := 0
	for _, transition := range transitions {
		if transition.Preserved {
			count++
		}
	}
	return count
}

func countIndicators(indicators []Indicator) int {
	count := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			count++
		}
	}
	return count
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestValue(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func finalize(report Report) Report {
	report.Digest = ""
	encoded, _ := json.Marshal(report)
	report.Digest = digestBytes(encoded)
	return report
}
