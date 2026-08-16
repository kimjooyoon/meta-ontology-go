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
	cases := []Case{
		bindingCase("case-01", ClosedSound, ReasonExactBinding, orderID, fileA, fileA, blobA, blobA, "Order", true, true, false),
		bindingCase("case-02", ClosedSound, ReasonRenameBinding, orderID, fileB, fileB, blobA, blobA, "PurchaseOrder", true, true, false),
		bindingCase("case-03", FailClosedUnsound, ReasonBlobWithoutID, orderID, fileA, fileB, blobA, blobB, "Order", true, true, false),
		bindingCase("case-04", UnknownFullSuiteRequired, ReasonSourceMapRegistry, orderID, fileA, fileA, blobA, blobA, "Order", false, false, true),
		missingBindingCase(),
		unknownGraphCase(),
		missedGraphCase(),
		soundnessCase("case-07", ReasonGlobalGuard, []string{"cmd-guard"}, guardCommands(), Authoritative, true),
		soundnessCase("case-08", ReasonSelectedDrift, []string{"cmd-impact"}, driftCommands(), Authoritative, true),
		soundnessCase("case-09", ReasonOmittedFailure, []string{"cmd-fail"}, omittedFailureCommands(), Authoritative, true),
		soundnessCase("case-10", ReasonNonAuthoritative, []string{"obl-candidate"}, nonAuthoritativeCommands(), Candidate, true),
		pathCase("case-11", ReasonDuplicateReceipt, "path://duplicate", true, false),
		pathCase("case-12", ReasonConflictingReceipt, "path://conflict", false, true),
		resourceCase(),
		externalCase("case-14", "external-authenticity", ExternalAssertion{Authenticity: false, Provider: true, Phase: true, Observer: true}),
		externalCase("case-15", "external-provider", ExternalAssertion{Authenticity: true, Provider: false, Phase: true, Observer: true}),
		externalCase("case-16", "external-phase", ExternalAssertion{Authenticity: true, Provider: true, Phase: false, Observer: true}),
		externalCase("case-17", "external-observer", ExternalAssertion{Authenticity: true, Provider: true, Phase: true, Observer: false}),
	}
	for index := range cases {
		cases[index].Baseline.WorkspaceRoot = "/workspace/fixture"
		cases[index].Baseline.SourcePath = "/workspace/fixture/" + cases[index].ID + ".go"
	}
	return cases
}

func missingBindingCase() Case {
	value := bindingCase("case-04-no-binding", UnknownFullSuiteRequired, ReasonBlobWithoutID, orderID, fileA, fileA, blobA, blobA, "Order", true, true, false)
	value.Expected.LocalizedIDs = []string{fileA}
	value.Baseline.DirectivePresent = false
	return value
}

func bindingCase(id string, state State, reason Reason, stableID, expectedFile, observedFile, expectedBlob, observedBlob, declaration string, registry, sourceMap, ambiguous bool) Case {
	return Case{ID: id, Expected: Expected{State: state, Reason: reason, LocalizedIDs: []string{stableID}}, Baseline: BaselineConfig{
		Subject: SubjectBinding, StableID: stableID, BoundID: stableID, ExpectedFile: expectedFile,
		ObservedFile: observedFile, ExpectedBlob: expectedBlob, ObservedBlob: observedBlob,
		DeclarationName: declaration, DirectivePresent: true,
		RegistryPresent: registry, SourceMapPresent: sourceMap, Ambiguous: ambiguous,
		FullCommands: 1, SelectedCommands: 1, ProvRecords: 1, ProvPaths: 1,
	}}
}

func unknownGraphCase() Case {
	return Case{ID: "case-05", Expected: Expected{State: UnknownFullSuiteRequired, Reason: ReasonUnknownGraph, LocalizedIDs: []string{"graph://unknown"}}, Baseline: BaselineConfig{
		Subject: SubjectGraph, UnknownIDs: []string{"graph://unknown"}, FullCommands: 3, SelectedCommands: 1, ProvRecords: 5, ProvPaths: 1,
	}}
}

func missedGraphCase() Case {
	return Case{ID: "case-06", Expected: Expected{State: FailClosedUnsound, Reason: ReasonMissedObligation, LocalizedIDs: []string{"billing://obligation/order"}}, Baseline: BaselineConfig{
		Subject: SubjectGraph, MissedIDs: []string{"billing://obligation/order"}, FullCommands: 3, SelectedCommands: 1, ProvRecords: 5, ProvPaths: 1,
	}}
}

func soundnessCase(id string, reason Reason, localized []string, commands []CommandAssertion, authority Authority, impacted bool) Case {
	obligationID := "obl-impact"
	if reason == ReasonNonAuthoritative {
		obligationID = "obl-candidate"
	}
	return Case{ID: id, Expected: Expected{State: expectedSoundnessState(reason), Reason: reason, LocalizedIDs: localized}, Baseline: BaselineConfig{
		Subject: SubjectSoundness, Commands: commands, Obligation: ObligationAssertion{ID: obligationID, Authority: authority, Impacted: impacted},
		External:     ExternalAssertion{Authenticity: true, Provider: true, Phase: true, Observer: true},
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

func pathCase(id string, reason Reason, pathID string, duplicate, conflict bool) Case {
	return Case{ID: id, Expected: Expected{State: FailClosedUnsound, Reason: reason, LocalizedIDs: []string{pathID}}, Baseline: BaselineConfig{
		Subject: SubjectPath, Path: PathAssertion{IDs: []string{pathID}, Duplicate: duplicate, Conflict: conflict}, FullCommands: 1, SelectedCommands: 1, ProvRecords: 4, ProvPaths: 2,
	}}
}

func resourceCase() Case {
	return Case{ID: "case-13", Expected: Expected{State: UnknownFullSuiteRequired, Reason: ReasonInvalidResource, LocalizedIDs: []string{"receipt-1"}}, Baseline: BaselineConfig{
		Subject: SubjectResource, Resource: ResourceAssertion{Overflow: true}, FullCommands: 1, SelectedCommands: 1, ProvRecords: 2, ProvPaths: 1,
	}}
}

func externalCase(id, localized string, external ExternalAssertion) Case {
	return Case{ID: id, Expected: Expected{State: UnknownFullSuiteRequired, Reason: ReasonExternalMissing, LocalizedIDs: []string{localized}}, Baseline: BaselineConfig{
		Subject: SubjectSoundness, External: external, Commands: guardCommands(), FullCommands: 2, SelectedCommands: 1, ProvRecords: 5, ProvPaths: 2,
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
