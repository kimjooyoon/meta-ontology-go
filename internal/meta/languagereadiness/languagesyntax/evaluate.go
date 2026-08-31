package languagesyntax

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

const conceptID = "language-syntax-roundtrip"

func Evaluate(repository fs.FS, headSHA string, registryRaw []byte,
	artifact languageconcept.Artifact) Report {
	source := Source{ExpectedHeadSHA: headSHA, ConceptArtifactDigest: artifact.ArtifactDigest,
		CatalogDigest: artifact.CatalogDigest, RegistryDigest: digestBytes(registryRaw),
		CorpusDigest: invalidDigest, ConceptRepositoryWrites: artifact.RepositoryWrites}
	report := Report{Schema: ReportSchema, Producer: "languagesyntax.Evaluate",
		Consumer: "self-improvement-cycle", MetaOperation: "prove-language-syntax-roundtrip",
		HeadSHA: headSHA, Source: source, RepositoryWrites: 0, MutationAuthorized: false}
	registry, registryErr := decodeRegistry(registryRaw)
	if registryErr != nil {
		report.Cases = unresolvedCases(report.Source)
		return finish(report)
	}
	report.Source.RegistryDigest = registryDigest()
	report.Source.PackageUnits = append([]PackageDefinition(nil), registry.PackageUnits...)
	report.Source.ConceptBound = conceptBound(repository, artifact)
	observed, observationErr := replay.Observe(repository)
	if observationErr == nil {
		report.Source.ObservationKnown = true
		report.Source.GoooFiles = observed
		report.Source.CorpusDigest = digestJSON(observed)
		report.Source.UnregisteredGooo, report.Source.MissingRegistered = compareCorpus(registry, observed)
	}
	if observationErr != nil || !report.Source.ConceptBound {
		report.Cases = unresolvedCases(report.Source)
		return finish(report)
	}
	for _, definition := range registry.Cases {
		evidence := replay.Execute(repository, definition.Path, definition.Kind, definition.ExpectedDiagnostic)
		if definition.ImplicitActivityPorts {
			evidence = replay.ExecuteWithImplicitActivityPorts(repository, definition.Path, definition.Kind, definition.ExpectedDiagnostic)
		} else if definition.EntityFields {
			evidence = replay.ExecuteWithEntityFieldsSupport(repository, definition.Path, definition.Kind, definition.ExpectedDiagnostic)
		}
		item := CaseResult{Definition: definition, Evidence: evidence}
		item.Status = caseStatus(item)
		item.EvidenceDigest = caseDigest(item, report.Source)
		report.Cases = append(report.Cases, item)
	}
	return finish(report)
}

func compareCorpus(registry Registry, observed []replay.FileObservation) ([]string, []string) {
	expected, present := map[string]bool{}, map[string]bool{}
	registered := registryPaths(registry)
	for _, path := range registered {
		expected[path] = true
	}
	extra := []string{}
	for _, file := range observed {
		present[file.Path] = true
		if !expected[file.Path] {
			extra = append(extra, file.Path)
		}
	}
	missing := []string{}
	for _, path := range registered {
		if !present[path] {
			missing = append(missing, path)
		}
	}
	return extra, missing
}
