package languagegointeroperation

import (
	"bytes"

	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
)

func executeGeneratorCase(definition Definition) CaseResult {
	model, found := generatorFixture(definition.Fixture)
	if !found {
		return finishCase(definition, rejectedEvidence("REGISTRY", "GENERATOR_FIXTURE_UNKNOWN"), false)
	}
	first, err := generator.Generate(model, nil)
	if err != nil {
		return finishCase(definition, rejectedEvidence("GENERATOR", "GENERATOR_REJECTED"), false)
	}
	replay, err := generator.Generate(model, first.Source)
	if err != nil {
		return finishCase(definition, rejectedEvidence("GENERATOR_REPLAY", "GENERATOR_REPLAY_REJECTED"), false)
	}
	evidence, failure := inspectPositive(first.Source, replay.Source)
	if failure != nil {
		return finishCase(definition, rejectedEvidence(failure.Stage, failure.Code), false)
	}
	bindGeneratorEvidence(&evidence, first, replay)
	satisfied := bytes.Equal(first.Source, replay.Source) && evidence.SourceMapMappings > 0
	satisfied = satisfied && evidence.SourceMapDigest == evidence.ReplaySourceMap
	return finishCase(definition, evidence, satisfied)
}

func bindGeneratorEvidence(evidence *Evidence, first, replay generator.Result) {
	evidence.SourceMapMappings = len(first.SourceMap.Mappings)
	evidence.SourceMapDigest = digestJSON(first.SourceMap)
	evidence.ReplaySourceMap = digestJSON(replay.SourceMap)
}
