package safeworkbinding

import "testing"

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

func TestCanonicalResult_InvalidEnumValues(t *testing.T) {
	_, ok := decisionField("decision", Decision(255))
	check(t, !ok, "invalid decision")
	_, ok = reasonField("reason", Reason(255))
	check(t, !ok, "invalid reason")
	_, ok = enforcementEffectField("effect", EnforcementEffect(255))
	check(t, !ok, "invalid effect")
	_, ok = reasonListField("faults", []Reason{Reason(255)})
	check(t, !ok, "invalid reason list")
	_, ok = resultFrame(ParseResult{Decision: Decision(255)})
	check(t, !ok, "invalid result")
}

func TestBindingDigest_ConstructionOrderAndListInvariance(t *testing.T) {
	binding := baseBindingForDigest()
	permuted := SafeWorkBinding{
		ToolchainOptionsDigest: binding.ToolchainOptionsDigest,
		RegistryDigest:         binding.RegistryDigest,
		PolicyDigest:           binding.PolicyDigest,
		SemanticSnapshotDigest: binding.SemanticSnapshotDigest,
		SourceSnapshotDigest:   binding.SourceSnapshotDigest,
		ObligationID:           binding.ObligationID,
		PathID:                 binding.PathID,
		TaskID:                 binding.TaskID,
		Schema:                 binding.Schema,
	}
	check(t, bindingDigest(binding) == bindingDigest(permuted), "binding construction order")
	nilResult := passResultForDigest()
	nilResult.Faults = nil
	emptyResult := passResultForDigest()
	nilDigest, nilOK := resultDigest(nilResult)
	emptyDigest, emptyOK := resultDigest(emptyResult)
	check(t, nilOK, "nil result list")
	check(t, emptyOK, "empty result list")
	check(t, nilDigest == emptyDigest, "result list invariance")
}

func TestBindingDigest_GovernedFieldMutations(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func(*SafeWorkBinding)
	}{
		{
			"schema",
			func(binding *SafeWorkBinding) { binding.Schema += "-v2" },
		},
		{
			"task",
			func(binding *SafeWorkBinding) { binding.TaskID += "-v2" },
		},
		{
			"path",
			func(binding *SafeWorkBinding) { binding.PathID += "-v2" },
		},
		{
			"obligation",
			func(binding *SafeWorkBinding) { binding.ObligationID += "-v2" },
		},
		{
			"source",
			func(binding *SafeWorkBinding) { binding.SourceSnapshotDigest += "-v2" },
		},
		{
			"semantic",
			func(binding *SafeWorkBinding) { binding.SemanticSnapshotDigest += "-v2" },
		},
		{
			"policy",
			func(binding *SafeWorkBinding) { binding.PolicyDigest += "-v2" },
		},
		{
			"registry",
			func(binding *SafeWorkBinding) { binding.RegistryDigest += "-v2" },
		},
		{
			"toolchain",
			func(binding *SafeWorkBinding) { binding.ToolchainOptionsDigest += "-v2" },
		},
	}
	base := bindingDigest(baseBindingForDigest())
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			binding := baseBindingForDigest()
			mutation.mutate(&binding)
			check(t, bindingDigest(binding) != base, "governed mutation")
		})
	}
}
