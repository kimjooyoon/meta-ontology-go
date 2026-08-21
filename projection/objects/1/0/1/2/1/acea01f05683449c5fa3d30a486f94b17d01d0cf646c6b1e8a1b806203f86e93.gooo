package safeworkbinding

import (
	"testing"
)

func TestBindingDigest_Base(t *testing.T) {
	binding := baseBindingForDigest()
	check(t, len(bindingFrame(binding)) == 772, "binding frame length")
	check(t, bindingDigest(binding) == digestBase, "binding digest")
}
func TestResultDigest_PASS(t *testing.T) {
	result := passResultForDigest()
	frame, ok := resultFrame(result)
	check(t, ok, "pass result frame")
	check(t, len(frame) == 255, "pass result frame length")
	digest, ok := resultDigest(result)
	check(t, ok, "pass result digest")
	check(t, digest == digestPass, "pass result vector")
}
func TestReplayDigest_PASS(t *testing.T) {
	binding := bindingDigest(baseBindingForDigest())
	result, ok := resultDigest(passResultForDigest())
	check(t, ok, "pass replay input")
	frame := replayFrame(binding, result)
	check(t, len(frame) == 252, "replay frame length")
	check(t, replayDigest(binding, result) == digestReplay, "replay vector")
	check(t, replayDigest(binding, result) == replayDigest(binding, result), "replay repeat")
}
func TestBindingDigest_PathIDMutation(t *testing.T) {
	binding := baseBindingForDigest()
	binding.PathID = "billing://path/pay-v2"
	check(t, bindingDigest(binding) == digestPath, "path mutation vector")
}
func TestBindingDigest_RegistryMutation(t *testing.T) {
	binding := baseBindingForDigest()
	binding.RegistryDigest = repeatedDigest('9')
	check(t, bindingDigest(binding) == digestRegistry, "registry mutation vector")
}
func TestCanonicalResult_MissingBindingDigestFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionUnknown, ReasonRequiredInputMissing, true), digestMissing)
}
func TestCanonicalResult_DuplicateKeyFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonDuplicateKey, false), digestDuplicate)
}
func TestCanonicalResult_ExpectedLabelFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonUnknownField, false), digestExpectedLabel)
}
func TestCanonicalResult_BindingDigestMismatchFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonBindingDigestMismatch, false), digestMismatch)
}
func TestCanonicalResult_BOMForbiddenMixedInvalidUTF8Fixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonBOMForbidden, false), digestBOMMixed)
}
func TestCanonicalResult_SelfExclusion(t *testing.T) {
	binding := baseBindingForDigest()
	wantBinding := bindingDigest(binding)
	binding.BindingDigest = "sha256:mutated"
	check(t, bindingDigest(binding) == wantBinding, "binding self exclusion")
	result := passResultForDigest()
	wantResult, ok := resultDigest(result)
	check(t, ok, "result self input")
	result.ResultDigest = "sha256:mutated"
	gotResult, ok := resultDigest(result)
	check(t, ok, "result self digest")
	check(t, gotResult == wantResult, "result self exclusion")
}
