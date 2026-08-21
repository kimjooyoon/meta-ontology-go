package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateSourceMapContext(input Input) *ConstructionError {
	if input.SourceMap.Before == nil || input.SourceMap.Head == nil {
		return unknownError(CodeMissingAuthority, "before and head source-map observations are required")
	}
	if len(input.SourceMap.CandidateBindings) != 0 {
		return unknownError(CodeCandidateBinding, "candidate source-map observations are not authoritative")
	}
	if len(input.SourceMap.DerivedBindings) != 0 {
		return unknownError(CodeDerivedBinding, "derived source-map observations are not authoritative")
	}
	if _, err := registrySurfaceIndex(input.Authority.Registry.Surfaces); err != nil {
		return constructionError(err)
	}
	return nil
}
func resolveSnapshots(input Input) (map[semantic.ID]resolved, map[semantic.ID]resolved, *ConstructionError) {
	surfaces, err := registrySurfaceIndex(input.Authority.Registry.Surfaces)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	beforeSnapshot, err := snapshotIndex(*input.Before)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	headSnapshot, err := snapshotIndex(*input.Head)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnregistered(beforeSnapshot, surfaces); err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnregistered(headSnapshot, surfaces); err != nil {
		return nil, nil, constructionError(err)
	}
	beforeBindings, err := observationIndex(input.SourceMap.Before, surfaces)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	headBindings, err := observationIndex(input.SourceMap.Head, surfaces)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnobservedBindings(beforeSnapshot, beforeBindings); err != nil {
		return nil, nil, constructionError(err)
	}
	if err := rejectUnobservedBindings(headSnapshot, headBindings); err != nil {
		return nil, nil, constructionError(err)
	}
	before, err := resolveSide(beforeSnapshot, beforeBindings, surfaces, false)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	head, err := resolveSide(headSnapshot, headBindings, surfaces, true)
	if err != nil {
		return nil, nil, constructionError(err)
	}
	if err := requireCoverage(surfaces, before, head); err != nil {
		return nil, nil, constructionError(err)
	}
	return before, head, nil
}
