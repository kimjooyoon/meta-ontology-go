package languagesemanticbinding

func validateSemantic(value semanticArtifact, head string, metrics []string) error {
	identity := value.Schema == SemanticSchema && value.Decision == "PASS" && value.Resolution == "EXACT"
	identity = identity && value.ReasonCode != "" && validDigest(value.ReportDigest)
	if err := require(identity, "semantic identity mismatch"); err != nil {
		return err
	}
	if err := require(value.RepositoryWrites == 0 && !value.MutationAuthorized, "semantic observer gained authority"); err != nil {
		return err
	}
	source := value.Source
	bound := source.ExpectedHeadSHA == head && source.ConceptID == ConceptID && source.MetaOperation == MetaOperation
	bound = bound && source.Producer == "languagesemantic.Evaluate" && source.Consumer == "self-improvement-cycle"
	bound = bound && source.ObservationKnown && source.ConceptBound && validDigest(source.RegistryDigest)
	if err := require(bound, "semantic source binding mismatch"); err != nil {
		return err
	}
	syntax := source.SyntaxSummary
	if err := require(syntax.Satisfied == syntaxCaseDenominator && syntax.Total == syntaxCaseDenominator && syntax.ValidCases == syntaxValidSourceDenominator, "syntax inheritance mismatch"); err != nil {
		return err
	}
	if err := require(syntax.InvalidCases == syntaxInvalidCaseDenominator && syntax.GoooLines == syntaxGoooLineDenominator, "syntax denominator mismatch"); err != nil {
		return err
	}
	if err := validateSemanticSummary(value.Summary); err != nil {
		return err
	}
	return validateSemanticEvidence(value, metrics)
}
