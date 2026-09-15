package languagegointeroperation

import "bytes"

func inspectPositive(source, replaySource []byte) (Evidence, *inspectionFailure) {
	first, failure := inspectSource(source)
	if failure != nil {
		return Evidence{}, failure
	}
	if len(replaySource) == 0 {
		replaySource = first.Canonical
	}
	replay, failure := inspectSource(replaySource)
	if failure != nil {
		return Evidence{}, failure
	}
	canonicalEqual := bytes.Equal(first.Canonical, replay.Canonical)
	typeEqual := first.TypesBound && replay.TypesBound && first.APIDigest == replay.APIDigest
	evidence := Evidence{ActualOutcome: "ACCEPT", SourceDigest: digestBytes(source),
		ReplaySourceDigest: digestBytes(replaySource), CanonicalDigest: digestBytes(first.Canonical),
		ReplayCanonical: digestBytes(replay.Canonical), APIDigest: first.APIDigest,
		ReplayAPIDigest: replay.APIDigest, ExportedObjects: first.ExportedObjects,
		GenericMethods: first.GenericMethods, AliasNodes: first.AliasNodes,
		ASTReifications: 2, CanonicalReplay: canonicalEqual, TypeIdentityReplay: typeEqual}
	return evidence, nil
}

func executeGo127Case(definition Definition) CaseResult {
	source, found := go127Fixture(definition.Fixture)
	if !found {
		return finishCase(definition, rejectedEvidence("REGISTRY", "GO_1_27_FIXTURE_UNKNOWN"), false)
	}
	evidence, failure := inspectPositive(source, nil)
	if failure != nil {
		return finishCase(definition, rejectedEvidence(failure.Stage, failure.Code), false)
	}
	satisfied := evidence.CanonicalReplay && evidence.TypeIdentityReplay
	return finishCase(definition, evidence, satisfied)
}
