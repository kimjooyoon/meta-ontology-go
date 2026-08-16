package metarecognition

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	orderID = "billing://order"
	fileA   = "order.go"
	fileB   = "renamed.go"
	blobA   = "blob-a"
	blobB   = "blob-b"
)

func Corpus() []Case {
	return []Case{
		bindingCase("case-01", ClosedSound, ReasonExactBinding, orderID, fileA, fileA, blobA, blobA, true, true, false),
		bindingCase("case-02", ClosedSound, ReasonRenameBinding, orderID, fileB, fileB, blobA, blobA, true, true, false),
		bindingCase("case-03", FailClosedUnsound, ReasonBlobWithoutID, orderID, fileA, fileB, blobA, blobB, true, true, false),
		bindingCase("case-04", UnknownFullSuiteRequired, ReasonSourceMapRegistry, orderID, fileA, fileA, blobA, blobA, false, false, true),
		graphCase(),
		soundnessCase("case-06", ReasonGlobalGuard, []string{"cmd-guard"}, guardCommands(), Authoritative, true),
		soundnessCase("case-07", ReasonSelectedDrift, []string{"cmd-impact"}, driftCommands(), Authoritative, true),
		soundnessCase("case-08", ReasonOmittedFailure, []string{"cmd-fail"}, omittedFailureCommands(), Authoritative, true),
		soundnessCase("case-09", ReasonNonAuthoritative, []string{"obl-candidate"}, nonAuthoritativeCommands(), Candidate, true),
		pathCase(),
		resourceCase(),
		externalCase(),
	}
}

func bindingCase(id string, state State, reason Reason, stableID, expectedFile, observedFile, expectedBlob, observedBlob string, registry, sourceMap, ambiguous bool) Case {
	return Case{ID: id, Expected: Expected{State: state, Reason: reason, LocalizedIDs: []string{stableID}}, Baseline: BaselineConfig{
		Subject: SubjectBinding, StableID: stableID, BoundID: stableID, ExpectedFile: expectedFile,
		ObservedFile: observedFile, ExpectedBlob: expectedBlob, ObservedBlob: observedBlob,
		RegistryPresent: registry, SourceMapPresent: sourceMap, Ambiguous: ambiguous,
		FullCommands: 1, SelectedCommands: 1, ProvRecords: 1, ProvPaths: 1,
	}}
}

func graphCase() Case {
	return Case{ID: "case-05", Expected: Expected{State: UnknownFullSuiteRequired, Reason: ReasonUnknownGraph, LocalizedIDs: []string{"graph://unknown"}}, Baseline: BaselineConfig{
		Subject: SubjectGraph, UnknownIDs: []string{"graph://unknown"}, MissedIDs: []string{"billing://obligation/order"},
		FullCommands: 2, SelectedCommands: 2, ProvRecords: 5, ProvPaths: 1,
	}}
}

func soundnessCase(id string, reason Reason, localized []string, commands []CommandAssertion, authority Authority, impacted bool) Case {
	obligationID := "obl-impact"
	if reason == ReasonNonAuthoritative {
		obligationID = "obl-candidate"
	}
	return Case{ID: id, Expected: Expected{State: expectedSoundnessState(reason), Reason: reason, LocalizedIDs: localized}, Baseline: BaselineConfig{
		Subject: SubjectSoundness, Commands: commands, Obligation: ObligationAssertion{ID: obligationID, Authority: authority, Impacted: impacted},
		FullCommands: 2, SelectedCommands: selectedCount(commands), ProvRecords: 5, ProvPaths: 2,
	}}
}

func expectedSoundnessState(reason Reason) State {
	if reason == ReasonNonAuthoritative {
		return UnknownFullSuiteRequired
	}
	return FailClosedUnsound
}

func command(id string, selected bool, full, chosen Status, fullDigest, chosenDigest string, guard, impacted, failure bool) CommandAssertion {
	return CommandAssertion{ID: id, Selected: selected, FullStatus: full, SelectedStatus: chosen, FullDigest: fullDigest, SelectedDigest: chosenDigest, GlobalGuard: guard, Impacted: impacted, FullFailure: failure}
}

func guardCommands() []CommandAssertion {
	return []CommandAssertion{command("cmd-guard", false, Pass, Pass, "guard-full", "guard-full", true, false, false), command("cmd-impact", true, Pass, Pass, "impact-full", "impact-full", false, true, false)}
}

func driftCommands() []CommandAssertion {
	return []CommandAssertion{command("cmd-guard", true, Pass, Pass, "guard-full", "guard-full", true, false, false), command("cmd-impact", true, Pass, Fail, "impact-full", "impact-selected", false, true, false)}
}

func omittedFailureCommands() []CommandAssertion {
	return []CommandAssertion{command("cmd-guard", true, Pass, Pass, "guard-full", "guard-full", true, false, false), command("cmd-fail", false, Fail, Pass, "fail-full", "fail-full", false, true, true)}
}

func nonAuthoritativeCommands() []CommandAssertion {
	return []CommandAssertion{command("cmd-guard", true, Pass, Pass, "guard-full", "guard-full", true, false, false), command("cmd-candidate", false, Pass, Pass, "candidate-full", "candidate-full", false, false, false)}
}

func pathCase() Case {
	return Case{ID: "case-10", Expected: Expected{State: FailClosedUnsound, Reason: ReasonDuplicateReceipt, LocalizedIDs: []string{"path://duplicate"}}, Baseline: BaselineConfig{
		Subject: SubjectPath, Path: PathAssertion{IDs: []string{"path://duplicate"}, Duplicate: true}, FullCommands: 1, SelectedCommands: 1, ProvRecords: 4, ProvPaths: 2,
	}}
}

func resourceCase() Case {
	return Case{ID: "case-11", Expected: Expected{State: UnknownFullSuiteRequired, Reason: ReasonInvalidResource, LocalizedIDs: []string{"receipt-1"}}, Baseline: BaselineConfig{
		Subject: SubjectResource, Resource: ResourceAssertion{Overflow: true}, FullCommands: 1, SelectedCommands: 1, ProvRecords: 2, ProvPaths: 1,
	}}
}

func externalCase() Case {
	return Case{ID: "case-12", Expected: Expected{State: UnknownFullSuiteRequired, Reason: ReasonExternalMissing, LocalizedIDs: []string{"external-input"}}, Baseline: BaselineConfig{
		Subject: SubjectSoundness, External: ExternalAssertion{Authenticity: true, Provider: false, Phase: true, Observer: false}, Commands: guardCommands(), FullCommands: 2, SelectedCommands: 1, ProvRecords: 5, ProvPaths: 2,
	}}
}

func selectedCount(commands []CommandAssertion) int {
	count := 0
	for _, value := range commands {
		if value.Selected {
			count++
		}
	}
	return count
}

func digest(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:])
}
