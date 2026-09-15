package metarecognition

import (
	"crypto/sha256"
	"encoding/hex"
)

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
