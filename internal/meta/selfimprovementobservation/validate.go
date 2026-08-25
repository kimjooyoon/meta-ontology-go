package selfimprovementobservation

type validation struct {
	SourceSchema, ExactHead, SourceDigest, Contract bool
	FixedDenominators, MinimalValueState            bool
	ValueWitnesses, CompilerWitnesses               bool
	ResourceWitnesses, MetaOperations               bool
	Proofs, Views, Counterexamples                   bool
	SourceEffects, ReadOnlyAuthority                 bool
}

func validateInputs(in Inputs, opts Options) validation {
	report := in.Report.Value
	return validation{
		SourceSchema:      report.Schema == "gooo/language-example-experiment-report/v2",
		ExactHead:         shaPattern.MatchString(opts.HeadSHA) && opts.SourceRunID > 0 && report.SubjectSHA == opts.HeadSHA && in.Contract.Value.CommitSHA == opts.HeadSHA,
		SourceDigest:      validSourceReportDigest(report) && digestPattern.MatchString(report.FactsDigest) && digestPattern.MatchString(in.Report.FileDigest) && digestPattern.MatchString(in.Counterexamples.FileDigest) && digestPattern.MatchString(in.Contract.FileDigest),
		Contract:          validContract(in.Contract.Value),
		FixedDenominators: validDenominators(report),
		MinimalValueState: report.Decision == "PASS" && report.Resolution == "EXACT" && report.Reason == "EXPERIMENT_CONTRACT_OBSERVED" && report.Interpretation == "MINIMAL_VALUE_OBSERVED" && report.ContractID != "",
		ValueWitnesses:    validValue(report.Summary.Value),
		CompilerWitnesses: validCompiler(report.Summary.Compiler),
		ResourceWitnesses: validResources(report.Summary.Resources),
		MetaOperations:    validSourceIndicators(report.Indicators),
		Proofs:            validSourceProofs(report),
		Views:             validSourceViews(report.Views),
		Counterexamples:   validCounterexamples(in.Counterexamples.Value),
		SourceEffects:     report.Summary.Effects.RepositoryWrites == 0 && !report.Summary.Effects.MutationAuthority,
		ReadOnlyAuthority: true,
	}
}

func validContract(contract ContractReport) bool {
	counts, observation := map[string]int{}, false
	for _, indicator := range contract.Indicators {
		counts[indicator.Route]++
		if indicator.Verdict != "PASS" {
			return false
		}
		observation = observation || indicator.ID == "coherence.read-only-language-observation"
	}
	return contract.Schema == "gooo/self-improvement-contract/v1" && contract.Status == "PASS" &&
		!contract.PromotionAuthorized && rawDigestPattern.MatchString(contract.SemanticHash) &&
		rawDigestPattern.MatchString(contract.RegistryDigest) && len(contract.Indicators) == 8 &&
		counts["FOUNDATION"] == 3 && counts["COHERENCE"] == 4 && counts["REGRESSION"] == 1 && observation
}
