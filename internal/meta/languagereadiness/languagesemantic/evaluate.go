package languagesemantic

import (
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
	"os"
	"path/filepath"
)

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
