package languagesourcebindingpromotion

import (
	"bytes"
	"reflect"
)

func Evaluate(input Input) Report {
	report := Report{Schema: ReportSchema, Scope: Scope, HeadSHA: input.HeadSHA,
		ContractDigest: digestJSON(input.Contract), PolicySourceDigest: digestBytes(input.PolicySource),
		PolicyArtifactDigest: digestBytes(input.PolicyArtifact), IndependenceDigest: digestJSON(input.Independence),
		NotClaimed: []string{"full compiler semantic correctness", "producer validator repaired",
			"independent toolchain implementation", "value-level computation", "external side effects"}}
	definitions := input.Contract.Cases
	if len(definitions) == 5 {
		report.Cases = []CaseResult{
			evaluateCase(definitions[0], input.Producer, input.Receipt, input.Oracle, input.HeadSHA, report.PolicyArtifactDigest),
			evaluateCase(definitions[1], input.Producer, input.Receipt, nil, input.HeadSHA, report.PolicyArtifactDigest),
			evaluateCase(definitions[2], input.Producer, input.Receipt, input.UnknownOracle, input.HeadSHA, report.PolicyArtifactDigest),
			evaluateCase(definitions[3], input.Producer, input.Receipt, input.MismatchedOracle, input.HeadSHA, report.PolicyArtifactDigest),
			evaluateCase(definitions[4], input.UnknownProducer, input.Receipt, input.Oracle, input.HeadSHA, report.PolicyArtifactDigest),
		}
	}
	report.Summary = summarize(report.Cases, input)
	report.Decision, report.Resolution, report.Reason = DecisionClosed, ResolutionInvariant, "SOURCE_BINDING_PROMOTION_CONTRACT_MISMATCH"
	if reflect.DeepEqual(input.Contract, CanonicalContract()) && validHead(input.HeadSHA) &&
		len(input.PolicySource) > 0 && len(input.PolicyArtifact) > 0 && bytes.Equal(input.PolicyArtifact, input.PolicyReplayArtifact) &&
		input.Independence.Schema == IndependenceSchema && input.Independence.ProducerDependencies == 0 &&
		report.Summary.CasesSatisfied == 5 {
		report.Decision, report.Resolution, report.Reason = DecisionPass, ResolutionExact, "SOURCE_BINDING_PROMOTION_CONTRACT_SATISFIED"
	}
	report.Indicators = indicators(report)
	report.Proofs = proofs(report)
	report.Digest = reportDigest(report)
	return report
}
