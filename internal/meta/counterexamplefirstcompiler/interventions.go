package counterexamplefirstcompiler

import (
	"fmt"

	cf "github.com/kimjooyoon/meta-ontology-go/internal/meta/counterexamplefirst"
)

const interventionSchema = "gooo/counterexample-first-interventions/v1"

// AnalyzeInterventions records two causal controls over the same raw corpus:
// changing the Gooo-declared policy must alter the decision evidence, while a
// comment-only source change must not.
func AnalyzeInterventions(before, semanticAfter, commentAfter []byte, corpus cf.ScenarioCorpus) (cf.InterventionReport, error) {
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
	report := cf.InterventionReport{Schema: interventionSchema}
	report.SemanticIntervention.Before = beforeSide
	report.SemanticIntervention.After = semanticSide
	report.SemanticIntervention.SemanticDigestEqual = beforeSide.SemanticDigest == semanticSide.SemanticDigest
	report.SemanticIntervention.DecisionChanged = beforeSide.Decision != semanticSide.Decision || beforeSide.Resolution != semanticSide.Resolution
	report.SemanticIntervention.CounterexampleChanged = beforeSide.FirstCounterexampleID != semanticSide.FirstCounterexampleID
	report.SemanticIntervention.ClaimTransitionChanged = beforeSide.ClaimTransitionDigest != semanticSide.ClaimTransitionDigest
	report.CommentOnlyIntervention.Before = beforeSide
	report.CommentOnlyIntervention.After = commentSide
	report.CommentOnlyIntervention.SemanticDigestEqual = beforeSide.SemanticDigest == commentSide.SemanticDigest
	report.CommentOnlyIntervention.DecisionEqual = beforeSide.Decision == commentSide.Decision && beforeSide.Resolution == commentSide.Resolution
	report.CommentOnlyIntervention.CounterexampleEqual = beforeSide.FirstCounterexampleID == commentSide.FirstCounterexampleID
	report.CommentOnlyIntervention.ClaimTransitionEqual = beforeSide.ClaimTransitionDigest == commentSide.ClaimTransitionDigest
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
	side := cf.InterventionSide{SourceDigest: cf.DigestBytes(source), SemanticDigest: program.SemanticDigest, Rule: rule}
	var transitions []cf.ClaimTransition
	for _, spec := range contract.Cases {
		for _, scenario := range corpus.Scenarios {
			if scenario.ID != spec.ID {
				continue
			}
			receipt := compileScenario(contract, "intervention", path, source, program, rule, spec, scenario)
			if side.Decision == "" {
				side.Decision, side.Resolution = receipt.Decision, receipt.Resolution
			}
			if receipt.Counterexample != nil {
				side.CounterexamplesObserved++
				if side.FirstCounterexampleID == "" {
					side.FirstCounterexampleID = receipt.Counterexample.ID
				}
			}
			transitions = append(transitions, receipt.ClaimTransitions...)
		}
	}
	side.ClaimTransitionDigest = cf.DigestValue(transitions)
	return side, nil
}
