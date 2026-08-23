package languagesemantic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type Input struct {
	Root               string
	ExpectedHeadSHA    string
	RegistryPath       string
	SyntaxArtifactPath string
}

type syntaxReceipt struct {
	Schema             string        `json:"schema"`
	Decision           string        `json:"decision"`
	Resolution         string        `json:"resolution"`
	Summary            SyntaxSummary `json:"summary"`
	Source             syntaxSource  `json:"source"`
	Cases              []syntaxCase  `json:"cases"`
	RepositoryWrites   int           `json:"repository_writes"`
	MutationAuthorized bool          `json:"mutation_authorized"`
	ReportDigest       string        `json:"report_digest"`
}

type syntaxSource struct {
	ExpectedHeadSHA  string     `json:"expected_head_sha"`
	ObservationKnown bool       `json:"observation_known"`
	ConceptBound     bool       `json:"concept_bound"`
	GoooFiles        []GoooFile `json:"gooo_files"`
}

type syntaxCase struct {
	Definition struct {
		ID string `json:"id"`
	} `json:"definition"`
	Evidence struct {
		ObservedDecision string   `json:"observed_decision"`
		Diagnostics      []string `json:"diagnostics"`
	} `json:"evidence"`
	Status string `json:"status"`
}

func Evaluate(input Input) (Report, error) {
	registry, registryRaw, err := LoadRegistry(input.RegistryPath)
	if err != nil {
		return Report{}, err
	}
	if err := validateHeadSHA(input.ExpectedHeadSHA); err != nil {
		return Report{}, err
	}
	syntaxRaw, err := os.ReadFile(input.SyntaxArtifactPath)
	if err != nil {
		return unresolvedReport(registry, input.ExpectedHeadSHA, digestBytes(registryRaw), "syntax artifact unavailable: "+err.Error()), nil
	}
	var upstream syntaxReceipt
	if err := json.Unmarshal(syntaxRaw, &upstream); err != nil {
		return unresolvedReport(registry, input.ExpectedHeadSHA, digestBytes(registryRaw), "syntax artifact invalid: "+err.Error()), nil
	}
	if err := validateSyntaxReceipt(upstream, input.ExpectedHeadSHA); err != nil {
		report := unresolvedReport(registry, input.ExpectedHeadSHA, digestBytes(registryRaw), err.Error())
		report.Source.SyntaxArtifactDigest = digestBytes(syntaxRaw)
		report.Source.SyntaxReportDigest = upstream.ReportDigest
		finalizeReport(&report)
		return report, nil
	}

	root, err := filepath.Abs(input.Root)
	if err != nil {
		return Report{}, err
	}
	results := make([]CaseResult, 0, len(registry.Cases))
	sourcePaths := sourceDefinitions(registry)
	observedPaths := make([]string, 0, len(upstream.Source.GoooFiles))
	for _, file := range upstream.Source.GoooFiles {
		observedPaths = append(observedPaths, file.Path)
	}
	unregistered, missing := setDifferences(observedPaths, sourcePaths)

	observations := make(map[string]replay.Observation, expectedSources)
	var anchor *replay.Observation
	for _, definition := range registry.Cases {
		if definition.Kind != CaseSource {
			continue
		}
		observation, observeErr := replay.Observe(root, definition.Path)
		result := CaseResult{Definition: definition}
		if observeErr != nil {
			result.Status = StatusUnresolved
			result.Evidence.Error = observeErr.Error()
		} else {
			observations[definition.ID] = observation
			copyOf := observation
			result.Evidence.Source = &copyOf
			if sourceSatisfied(observation) {
				result.Status = StatusSatisfied
			} else {
				result.Status = StatusNotSatisfied
			}
			if anchor == nil && observation.DeterministicFacts > 0 {
				candidate := observation
				anchor = &candidate
			}
		}
		result.Digest = caseDigest(result)
		results = append(results, result)
	}

	lawObservation, lawErr := replay.LawObservation{}, error(nil)
	if anchor == nil {
		lawErr = fmt.Errorf("no source model contains a deterministic fact")
	} else {
		lawObservation, lawErr = replay.ObserveLaws(anchor.Path, anchor.IR)
	}
	upstreamCases := make(map[string]syntaxCase, len(upstream.Cases))
	for _, item := range upstream.Cases {
		upstreamCases[item.Definition.ID] = item
	}
	for _, definition := range registry.Cases {
		if definition.Kind == CaseSource {
			continue
		}
		result := CaseResult{Definition: definition}
		switch definition.Kind {
		case CaseLaw:
			if lawErr != nil {
				result.Status = StatusUnresolved
				result.Evidence.Error = lawErr.Error()
			} else {
				satisfied := lawSatisfied(definition.Law, lawObservation)
				result.Evidence.Law = &LawEvidence{Law: definition.Law, Satisfied: satisfied, Observation: lawObservation}
				if satisfied {
					result.Status = StatusSatisfied
				} else {
					result.Status = StatusNotSatisfied
				}
			}
		case CaseUpstreamRejection:
			item, ok := upstreamCases[definition.UpstreamCase]
			if !ok {
				result.Status = StatusUnresolved
				result.Evidence.Error = "upstream syntax case is missing"
			} else {
				evidence := &UpstreamEvidence{CaseID: definition.UpstreamCase, ObservedDecision: item.Evidence.ObservedDecision, Diagnostics: append([]string(nil), item.Evidence.Diagnostics...)}
				result.Evidence.Upstream = evidence
				if item.Status == "SATISFIED" && item.Evidence.ObservedDecision == "FAIL_CLOSED" {
					result.Status = StatusSatisfied
				} else {
					result.Status = StatusNotSatisfied
				}
			}
		}
		result.Digest = caseDigest(result)
		results = append(results, result)
	}

	report := buildReport(registry, results, Source{
		ExpectedHeadSHA:      input.ExpectedHeadSHA,
		ConceptID:            ConceptID,
		Producer:             "languagesemantic.Evaluate",
		Consumer:             "self-improvement-cycle",
		MetaOperation:        "prove-staged-semantic-model",
		RegistryDigest:       digestBytes(registryRaw),
		SyntaxArtifactDigest: digestBytes(syntaxRaw),
		SyntaxReportDigest:   upstream.ReportDigest,
		SyntaxSummary:        upstream.Summary,
		GoooFiles:            append([]GoooFile(nil), upstream.Source.GoooFiles...),
		ObservationKnown:     true,
		ConceptBound:         true,
	}, len(unregistered), len(missing), boolInt(len(unregistered) > 0 || len(missing) > 0))
	return report, nil
}

func validateSyntaxReceipt(receipt syntaxReceipt, expectedHead string) error {
	if receipt.Schema != "gooo/language-syntax-roundtrip/v1" {
		return fmt.Errorf("syntax evidence schema is unknown")
	}
	if receipt.Source.ExpectedHeadSHA != expectedHead {
		return fmt.Errorf("syntax evidence head does not match the semantic subject")
	}
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" {
		return fmt.Errorf("syntax evidence decision is not explicit PASS / EXACT")
	}
	if !receipt.Source.ObservationKnown || !receipt.Source.ConceptBound {
		return fmt.Errorf("syntax evidence is not dynamically bound")
	}
	if receipt.Summary.Satisfied != 15 || receipt.Summary.Total != 15 || receipt.Summary.ValidCases != 13 || receipt.Summary.InvalidCases != 2 || receipt.Summary.GoooLines != 174 {
		return fmt.Errorf("syntax evidence denominator does not match 15 cases / 13 files / 174 lines")
	}
	if len(receipt.Source.GoooFiles) != expectedSources {
		return fmt.Errorf("syntax evidence contains %d Gooo files, want %d", len(receipt.Source.GoooFiles), expectedSources)
	}
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthorized {
		return fmt.Errorf("syntax evidence crossed the read-only effect boundary")
	}
	return nil
}

func validateHeadSHA(value string) error {
	if len(value) != 40 {
		return fmt.Errorf("head SHA must contain 40 hexadecimal characters")
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return fmt.Errorf("head SHA must be lowercase hexadecimal")
		}
	}
	return nil
}

func sourceSatisfied(observation replay.Observation) bool {
	return observation.Normalized && observation.CanonicalReplay && observation.SemanticReplay && observation.ProvenanceReplay && observation.EvidenceReplay &&
		slices.Equal(observation.Stages, replay.ExpectedStages) && observation.Effects.Writes == 0 && observation.Effects.Network == 0 && observation.Effects.Processes == 0
}

func lawSatisfied(law string, observation replay.LawObservation) bool {
	switch law {
	case "PRESENTATION_INVARIANCE":
		return observation.PresentationChanged && observation.PresentationInvariant
	case "CANDIDATE_NON_AUTHORITY":
		return observation.CandidateRecorded && observation.CandidateCanonicalChanged && observation.CandidateNonAuthoritative
	case "DETERMINISTIC_AUTHORITY":
		return observation.DeterministicRecorded && observation.DeterministicCanonicalChanged && observation.DeterministicAuthoritative
	default:
		return false
	}
}

func sourceDefinitions(registry Registry) []string {
	paths := make([]string, 0, expectedSources)
	for _, definition := range registry.Cases {
		if definition.Kind == CaseSource {
			paths = append(paths, filepath.ToSlash(filepath.Clean(definition.Path)))
		}
	}
	return paths
}

func setDifferences(observed, registered []string) ([]string, []string) {
	observedSet, registeredSet := make(map[string]struct{}), make(map[string]struct{})
	for _, value := range observed {
		observedSet[filepath.ToSlash(filepath.Clean(value))] = struct{}{}
	}
	for _, value := range registered {
		registeredSet[filepath.ToSlash(filepath.Clean(value))] = struct{}{}
	}
	var unregistered, missing []string
	for value := range observedSet {
		if _, ok := registeredSet[value]; !ok {
			unregistered = append(unregistered, value)
		}
	}
	for value := range registeredSet {
		if _, ok := observedSet[value]; !ok {
			missing = append(missing, value)
		}
	}
	return unregistered, missing
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func unresolvedReport(registry Registry, head, registryDigest, reason string) Report {
	results := make([]CaseResult, 0, len(registry.Cases))
	for _, definition := range registry.Cases {
		result := CaseResult{Definition: definition, Status: StatusUnresolved, Evidence: CaseEvidence{Error: reason}}
		result.Digest = caseDigest(result)
		results = append(results, result)
	}
	return buildReport(registry, results, Source{
		ExpectedHeadSHA:  head,
		ConceptID:        ConceptID,
		Producer:         "languagesemantic.Evaluate",
		Consumer:         "self-improvement-cycle",
		MetaOperation:    "prove-staged-semantic-model",
		RegistryDigest:   registryDigest,
		ObservationKnown: false,
		ConceptBound:     false,
	}, 0, expectedSources, 1)
}

func semanticHash(value string) string {
	return semantic.StableHashString(value)
}
