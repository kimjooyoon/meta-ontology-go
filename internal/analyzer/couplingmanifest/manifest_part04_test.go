package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

func cloneInput(input Input) Input {
	copy := input
	copy.Authority.Registry.Surfaces = append([]detector.Surface(nil), input.Authority.Registry.Surfaces...)
	copy.SourceMap.Before = append([]SourceMapObservation(nil), input.SourceMap.Before...)
	copy.SourceMap.Head = append([]SourceMapObservation(nil), input.SourceMap.Head...)
	copy.SourceMap.CandidateBindings = append([]SourceMapObservation(nil), input.SourceMap.CandidateBindings...)
	copy.SourceMap.DerivedBindings = append([]SourceMapObservation(nil), input.SourceMap.DerivedBindings...)
	return copy
}
