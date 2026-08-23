package languagesemanticbinding

func validateConcept(value conceptArtifact) (conceptDefinition, error) {
	if err := require(value.Schema == ConceptSchema && value.Decision == "PASS", "concept artifact rejected"); err != nil {
		return conceptDefinition{}, err
	}
	if err := require(value.ReplayEqual && value.RepositoryWrites == 0 && validDigest(value.ArtifactDigest), "concept replay mismatch"); err != nil {
		return conceptDefinition{}, err
	}
	summary := value.Report.Summary
	counts := summary.Concepts == 15 && summary.CodeBound == 15 && summary.UseCaseBound == 15
	counts = counts && summary.MetricBound == 15 && summary.Operating == 13 && summary.Conformed == 2
	guards := summary.Unbound == 0 && summary.UnverifiedNoveltyClaims == 0 && summary.RepositoryWrites == 0
	if err := require(value.Report.Decision == "PASS" && counts && guards, "concept catalog mismatch"); err != nil {
		return conceptDefinition{}, err
	}
	for _, concept := range value.Report.Concepts {
		if concept.ID != ConceptID {
			continue
		}
		valid := concept.MetaOperation == MetaOperation && concept.Stage == "OPERATING" && !concept.NoveltyClaim
		valid = valid && len(concept.CodeBindings) == 6 && len(concept.MetricBindings) == 19
		valid = valid && contains(concept.CodeBindings, "internal/meta/languagereadiness/languagesemantic")
		valid = valid && len(concept.UseCases) == 1
		if len(concept.UseCases) == 1 {
			valid = valid && concept.UseCases[0].ID == "staged-semantic-authority-replay"
			valid = valid && concept.UseCases[0].ExpectedOutcome == "IMPROVED_14_TO_15_OF_24_WITH_18_OF_18_CASES_AND_ZERO_EFFECTS"
		}
		if err := require(valid, "semantic concept binding mismatch"); err != nil {
			return conceptDefinition{}, err
		}
		return concept, nil
	}
	return conceptDefinition{}, require(false, "semantic concept missing")
}
