package counterexamplefirstcompiler

import (
	"fmt"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
)

const interventionSchema = "gooo/counterexample-first-interventions/v2"

// AnalyzeInterventions records causal controls over the same raw corpus. A
// semantic rule intervention and a graph-authority intervention must alter
// per-case evidence; a comment-only source change must preserve it.
func AnalyzeInterventions(before, semanticAfter, commentAfter, metaAfter []byte, corpus cf.ScenarioCorpus) (cf.InterventionReport, error) {
	contract := cf.CanonicalContract()
	beforeSide, err := interventionSide("before.gooo", before, corpus, contract)
	if err != nil {
		return cf.InterventionReport{}, err
	}
	semanticSide, err := interventionSide("semantic-intervention.gooo", semanticAfter, corpus, contract)
	if err != nil {
		return cf.InterventionReport{}, err
	}
	commentSide, err := interventionSide("comment-only-intervention.gooo", commentAfter, corpus, contract)
	if err != nil {
		return cf.InterventionReport{}, err
	}
	metaSide, err := interventionSide("meta-operation-intervention.gooo", metaAfter, corpus, contract)
	if err != nil {
		return cf.InterventionReport{}, err
	}
	report := cf.InterventionReport{Schema: interventionSchema}
	report.SemanticIntervention = compareSides(beforeSide, semanticSide, false)
	report.CommentOnlyIntervention = compareSides(beforeSide, commentSide, true)
	report.MetaOperationIntervention = compareSides(beforeSide, metaSide, false)
	report.Digest = cf.DigestValue(report)
	return report, nil
}

func interventionSide(path string, source []byte, corpus cf.ScenarioCorpus, contract cf.Contract) (cf.InterventionSide, error) {
	program, rule, err := observeProgram(path, source)
	if err != nil || !program.ParseOK || !program.LowerOK {
		if err != nil {
			return cf.InterventionSide{}, err
		}
		return cf.InterventionSide{}, fmt.Errorf("intervention source was not observed")
	}
	side := cf.InterventionSide{SourceDigest: cf.DigestBytes(source), SemanticDigest: program.SemanticDigest, Rule: rule, MetaOperationAuthorized: program.MetaOperation.Authorized}
	byID := make(map[string]cf.Scenario, len(corpus.Scenarios))
	for _, scenario := range corpus.Scenarios {
		byID[scenario.ID] = scenario
	}
	for _, spec := range contract.Cases {
		scenario, ok := byID[spec.ID]
		predicateSpec, predicateOK := predicateByID(contract.Predicates, spec.PredicateID)
		if !ok || !predicateOK {
			return cf.InterventionSide{}, fmt.Errorf("intervention scenario missing: %s", spec.ID)
		}
		receipt := compileScenario(contract, "intervention", path, source, program, rule, predicateSpec, spec, scenario)
		side.Cases = append(side.Cases, cf.CaseIntervention{
			CaseID: spec.ID, ClaimID: receipt.ClaimID, PropositionDigest: receipt.PropositionDigest, PredicateID: receipt.PredicateID,
			TargetOperation: spec.MetaOperation, SourceDigest: receipt.CandidateObservation.SourceDigest, SemanticDigest: receipt.CandidateObservation.SemanticDigest,
			Decision: receipt.Decision, Resolution: receipt.Resolution, TransitionDigest: cf.DigestValue(receipt.ClaimTransitions),
			CounterexampleID: counterexampleID(receipt), MetaOperationAuthorized: receipt.ProgramMetaOperation.Authorized,
		})
	}
	side.CaseTransitionDigest = cf.DigestValue(side.Cases)
	return side, nil
}

func counterexampleID(receipt cf.DecisionReceipt) string {
	if receipt.Counterexample == nil {
		return ""
	}
	return receipt.Counterexample.ID
}

func compareSides(before, after cf.InterventionSide, equalExpected bool) cf.InterventionComparison {
	comparison := cf.InterventionComparison{Before: before, After: after, SemanticDigestEqual: before.SemanticDigest == after.SemanticDigest, AllCasesAddressed: len(before.Cases) == len(after.Cases) && len(before.Cases) == cf.CaseCount}
	for index := range before.Cases {
		if index >= len(after.Cases) {
			comparison.AllCasesAddressed = false
			continue
		}
		left, right := before.Cases[index], after.Cases[index]
		if left.CaseID != right.CaseID || left.ClaimID != right.ClaimID || left.PropositionDigest != right.PropositionDigest || left.PredicateID != right.PredicateID || left.TargetOperation != right.TargetOperation {
			comparison.AllCasesAddressed = false
		}
		if left.Decision != right.Decision || left.Resolution != right.Resolution {
			comparison.DecisionChanged = true
		}
		if left.CounterexampleID != right.CounterexampleID {
			comparison.CounterexampleChanged = true
		}
		if left.TransitionDigest != right.TransitionDigest {
			comparison.ClaimTransitionChanged = true
		}
	}
	if equalExpected {
		comparison.DecisionChanged = !decisionEqual(comparison.Before, comparison.After)
		comparison.CounterexampleChanged = !counterexampleEqual(comparison.Before, comparison.After)
		comparison.ClaimTransitionChanged = !claimTransitionEqual(comparison.Before, comparison.After)
	}
	return comparison
}

func decisionEqual(before, after cf.InterventionSide) bool {
	if len(before.Cases) != len(after.Cases) {
		return false
	}
	for index := range before.Cases {
		if before.Cases[index].Decision != after.Cases[index].Decision || before.Cases[index].Resolution != after.Cases[index].Resolution {
			return false
		}
	}
	return true
}

func counterexampleEqual(before, after cf.InterventionSide) bool {
	if len(before.Cases) != len(after.Cases) {
		return false
	}
	for index := range before.Cases {
		if before.Cases[index].CounterexampleID != after.Cases[index].CounterexampleID {
			return false
		}
	}
	return true
}

func claimTransitionEqual(before, after cf.InterventionSide) bool {
	if len(before.Cases) != len(after.Cases) {
		return false
	}
	for index := range before.Cases {
		if before.Cases[index].TransitionDigest != after.Cases[index].TransitionDigest {
			return false
		}
	}
	return true
}
