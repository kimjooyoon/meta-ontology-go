package selfimprovementobservation

type projection struct {
	Observation Observation
	Validation  validation
}

func project(in Inputs, opts Options) projection {
	report, counterexamples, contract := in.Report.Value, in.Counterexamples.Value, in.Contract.Value
	artifacts := []ArtifactRef{
		{Kind: "language-experiment-report", Schema: report.Schema, FileDigest: in.Report.FileDigest, SemanticDigest: report.Digest, Decision: report.Decision},
		{Kind: "language-experiment-counterexamples", Schema: counterexamples.Schema, FileDigest: in.Counterexamples.FileDigest, SemanticDigest: digestJSON(counterexamples), Decision: counterexampleDecision(counterexamples)},
		{Kind: "gooo-self-improvement-contract", Schema: contract.Schema, FileDigest: in.Contract.FileDigest, SemanticDigest: contract.SemanticHash, Decision: contract.Status},
	}
	inputDigest := digestJSON(struct {
		HeadSHA string
		RunID   int64
		Inputs  []ArtifactRef
	}{opts.HeadSHA, opts.SourceRunID, artifacts})
	return projection{Observation: Observation{
		Schema: observationSchema, Metaprogram: metaprogram, SubjectSHA: opts.HeadSHA,
		SourceWorkflowRunID: opts.SourceRunID, ContractID: report.ContractID,
		Summary: ObservationSummary{
			SourceCoordinates: report.Summary.Coordinates,
			Counterexamples: CountSummary{Satisfied: counterexamples.Satisfied, Total: counterexamples.Total, BasisPoints: basisPoints(counterexamples.Satisfied, counterexamples.Total)},
			GoooDefinitionFiles: report.Summary.Compiler.GoooFiles, GoDefinitionFiles: report.Summary.Compiler.GoFiles,
			ResourceSamples: report.Summary.Resources.ValidSamples, MaxWallMS: report.Summary.Resources.MaxWallMS,
			MaxRSSKiB: report.Summary.Resources.MaxRSSKiB, BinaryBytes: report.Summary.Resources.BinaryBytes,
			CandidateCount: 0,
		},
		Authority: Authority{}, Artifacts: artifacts, InputDigest: inputDigest,
		Indicators: []Indicator{}, Views: []View{}, Proofs: []Proof{}, NotClaimed: observationNonClaims(),
	}, Validation: validateInputs(in, opts)}
}

func counterexampleDecision(report CounterexampleReport) string {
	if validCounterexamples(report) {
		return "PASS"
	}
	return "FAIL_CLOSED"
}
