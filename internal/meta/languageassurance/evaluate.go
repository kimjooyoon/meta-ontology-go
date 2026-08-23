package languageassurance

import (
	"slices"
	"fmt"
	"sort"
)

func Evaluate(subjectSHA string, transaction Transaction) (Report, error) {
	if err := validateInput(subjectSHA, transaction); err != nil {
		return Report{}, err
	}
	definitions := Denominator()
	operations := CanonicalMetaOperations()
	obligations, operating := observeObligations(definitions)
	findings := append(detectSelfMinting(transaction), detectRoleConflicts(transaction)...)
	findings = append(findings, detectUnknownLaundering(transaction)...)
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].MetricID+findings[i].PathID < findings[j].MetricID+findings[j].PathID
	})

	authorityObserved := len(transaction.AuthorityRoutes) > 0
	rolesObserved := len(transaction.RoleBindings) > 0
	decisionsObserved := len(transaction.DecisionTransitions) > 0
	evidenceObserved := boolInt(authorityObserved) + boolInt(rolesObserved) + boolInt(decisionsObserved)
	selfMinting := countFindings(findings, MetricSelfMinting)
	roleConflicts := countFindings(findings, MetricRoleConflict)
	unknownLaundering := countFindings(findings, MetricUnknownLaundering)
	unknownTop := countUnknownTop(transaction.DecisionTransitions)

	summary := Summary{
		DenominatorTotal:          len(definitions),
		Operating:                 operating,
		NotImplemented:            len(definitions) - operating,
		ImplementationCoverageBPS: operating * 10000 / len(definitions),
		EvidenceGroupsObserved:    evidenceObserved,
		EvidenceGroupsTotal:       3,
		EvidenceCoverageBPS:       evidenceObserved * 10000 / 3,
		SelfMintingPaths:          observedValue(authorityObserved, selfMinting),
		RoleConflictPaths:         observedValue(rolesObserved, roleConflicts),
		UnknownLaunderingPaths:    observedValue(decisionsObserved, unknownLaundering),
		UnknownTopDecisions:       observedValue(decisionsObserved, unknownTop),
		RepositoryWrites:          0,
	}
	summary.UnresolvedIndicators = unresolved(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths)
	summary.ViolatedGuardrails = positive(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths)
	decision, reason, resolution := decide(summary)
	indicators := buildIndicators(summary)
	report := Report{
		Schema: ReportSchema, SubjectSHA: subjectSHA,
		TransactionDigest: digest(transaction), DenominatorID: DenominatorID,
		DenominatorDigest: digest(definitions), AssuranceDecision: AssurancePartial,
		CandidateDecision: decision, CandidateReason: reason, CandidateResolution: resolution,
		Denominator: definitions, Obligations: obligations, MetaOperations: operations,
		RoleConflictPairs: RoleConflictPairs(), UnknownLaunderingOutputs: UnknownLaunderingOutputs(),
		Transaction: transaction, Findings: findings, Indicators: indicators, Summary: summary,
	}
	seal(&report)
	return report, nil
}

func observeObligations(definitions []ObligationDefinition) ([]ObligationObservation, int) {
	observations := make([]ObligationObservation, 0, len(definitions))
	operating := 0
	for _, definition := range definitions {
		observation := ObligationObservation{MetricID: definition.MetricID, Status: "NOT_IMPLEMENTED", Resolution: ResolutionNone}
		if operation, ok := operatingOperations[definition.MetricID]; ok {
			observation.Status, observation.Resolution, observation.MetaOperation = "OPERATING", ResolutionExact, operation
			operating++
		}
		observations = append(observations, observation)
	}
	return observations, operating
}

func detectSelfMinting(transaction Transaction) []Finding {
	var findings []Finding
	for _, route := range transaction.AuthorityRoutes {
		if route.AuthoredBy == route.PromotedBy {
			findings = append(findings, Finding{MetricID: MetricSelfMinting, PathID: "authority/" + route.RuleID + "/" + route.AuthoredBy, Principal: route.AuthoredBy, RuleID: route.RuleID})
		}
	}
	return findings
}

func detectRoleConflicts(transaction Transaction) []Finding {
	var findings []Finding
	for _, binding := range transaction.RoleBindings {
		roles := roleSet(binding.Roles)
		for _, pair := range conflictPairs {
			if roles[pair.Left] && roles[pair.Right] {
				path := fmt.Sprintf("role/%s/%s+%s", binding.Principal, pair.Left, pair.Right)
				findings = append(findings, Finding{MetricID: MetricRoleConflict, PathID: path, Principal: binding.Principal, Roles: []Role{pair.Left, pair.Right}})
			}
		}
	}
	return findings
}

func detectUnknownLaundering(transaction Transaction) []Finding {
	var findings []Finding
	for _, transition := range transaction.DecisionTransitions {
		if transition.Input == DecisionUnknown && isLaunderingOutput(transition.Output) {
			path := fmt.Sprintf("decision/%s/%s->%s", transition.ID, transition.Input, transition.Output)
			findings = append(findings, Finding{MetricID: MetricUnknownLaundering, PathID: path, DecisionID: transition.ID, Input: transition.Input, Output: transition.Output})
		}
	}
	return findings
}

func buildIndicators(summary Summary) []Indicator {
	coverage := summary.ImplementationCoverageBPS
	evidence := summary.EvidenceCoverageBPS
	return []Indicator{
		indicator("gooo.metric.assurance.implementation-coverage-bps.v1", ClassOutcome, ProofCoherence, "freeze-assurance-denominator", &coverage, 10000, "basis_points", RelationGreaterOrEqual),
		indicator("gooo.metric.assurance.transaction-evidence-coverage-bps.v1", ClassDriver, ProofFoundation, "observe-transaction-evidence", &evidence, 10000, "basis_points", RelationGreaterOrEqual),
		indicator(MetricSelfMinting, ClassGuardrail, ProofFoundation, "detect-self-minting-paths", summary.SelfMintingPaths, 0, "paths", RelationLessOrEqual),
		indicator(MetricRoleConflict, ClassGuardrail, ProofCoherence, "detect-role-conflict-paths", summary.RoleConflictPaths, 0, "paths", RelationLessOrEqual),
		indicator(MetricUnknownLaundering, ClassGuardrail, ProofRegression, "detect-unknown-laundering", summary.UnknownLaunderingPaths, 0, "paths", RelationLessOrEqual),
	}
}

func indicator(metricID string, class IndicatorClass, proof ProofChoice, operation string, value *int, target int, unit string, relation Relation) Indicator {
	resolution := ResolutionExact
	if value == nil {
		resolution = ResolutionUnknown
	}
	return Indicator{MetricID: metricID, Class: class, ProofChoice: proof, Producer: Producer,
		Consumer: Consumer, MetaOperation: operation, Value: value, Target: target, Unit: unit,
		Relation: relation, Resolution: resolution, Satisfied: satisfies(value, target, relation)}
}

func decide(summary Summary) (string, string, Resolution) {
	if summary.UnresolvedIndicators > 0 || summary.EvidenceGroupsObserved != summary.EvidenceGroupsTotal {
		return CandidateFailClosed, ReasonEvidenceUnknown, ResolutionUnknown
	}
	if summary.UnknownTopDecisions != nil && *summary.UnknownTopDecisions > 0 {
		return CandidateFailClosed, ReasonTopDecisionUnknown, ResolutionInvariantOnly
	}
	if summary.ViolatedGuardrails > 0 {
		return CandidateBlock, ReasonGovernanceViolation, ResolutionExact
	}
	return CandidateAllowLimited, ReasonBoundaryClear, ResolutionExact
}

func countFindings(findings []Finding, metricID string) int {
	count := 0
	for _, finding := range findings {
		if finding.MetricID == metricID {
			count++
		}
	}
	return count
}

func countUnknownTop(transitions []DecisionTransition) int {
	count := 0
	for _, transition := range transitions {
		if transition.Input == DecisionUnknown {
			count++
		}
	}
	return count
}

func isLaunderingOutput(output Decision) bool {
	return slices.Contains(launderingOutputs, output)
}

func observedValue(observed bool, value int) *int {
	if !observed {
		return nil
	}
	return &value
}

func unresolved(values ...*int) int {
	count := 0
	for _, value := range values {
		if value == nil {
			count++
		}
	}
	return count
}

func positive(values ...*int) int {
	count := 0
	for _, value := range values {
		if value != nil && *value > 0 {
			count++
		}
	}
	return count
}

func satisfies(value *int, target int, relation Relation) bool {
	if value == nil {
		return false
	}
	if relation == RelationGreaterOrEqual {
		return *value >= target
	}
	return *value <= target
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
