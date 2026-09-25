package safeworkbinding

import (
	"testing"
)

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
