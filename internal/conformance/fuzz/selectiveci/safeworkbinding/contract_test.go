package safeworkbinding

import (
	"encoding/hex"
	"reflect"
	"testing"
)

const (
	ones       = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	twos       = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
	threes     = "sha256:3333333333333333333333333333333333333333333333333333333333333333"
	fours      = "sha256:4444444444444444444444444444444444444444444444444444444444444444"
	fives      = "sha256:5555555555555555555555555555555555555555555555555555555555555555"
	dBind      = "sha256:dc6dbe157ede5924b61676bfdcd4151cd6f73a51b7eefda674cca3d6d169a5cb"
	dPath      = "sha256:c262c10f1652d0d3fd5fa605166d565a44261b834ffcffe876f2bb785f4bcd51"
	dReg       = "sha256:23477b01e3d477d6ba5702adca8ad77a2cd643748b29e20ac4a7fda0b020673e"
	dResult    = "sha256:b076bd08e4b82b4b2aeb78f3f8a3e12931ae8f8a01f734cb2a97bca15753a342"
	dReplay    = "sha256:4564f79ab93fadeb3221f837e38dfdce43e1c653a89086283aaceb09d923080d"
	dMissing   = "sha256:3b4256d3ab6d60db4caa093cc8306819c9a6fb9fb987927fab8bff841a0690c8"
	dDuplicate = "sha256:bac3126c6e354dfb38abf69508605e50226f186d25b7381f8eb3215f33e7fda0"
	dUnknown   = "sha256:1ada140e4d914ab1ab15570deb11e423a7e454b81c2bbf84512454b642ecfa02"
	dBOM       = "sha256:49d26986dde23834ba79d1064aa577860dfe47a79bff2ab698d385660db1ac8c"
	dMismatch  = "sha256:c61d5823d85939c5f7645d6f2fe2049fe724bfec3fc12c011e9532ccc41c6b10"
)

type faultVector struct {
	name   string
	result ParseResult
	length int
	digest Digest
}

func assert(t *testing.T, ok bool, message string) {
	t.Helper()
	if !ok {
		t.Fatal(message)
	}
}

func baseBinding() SafeWorkBinding {
	return SafeWorkBinding{
		Schema: SafeWorkBindingSchemaV1, TaskID: "billing://task/pay",
		PathID: "billing://path/pay", ObligationID: "billing://obligation/pay",
		SourceSnapshotDigest: ones, SemanticSnapshotDigest: twos,
		PolicyDigest: threes, RegistryDigest: fours, ToolchainOptionsDigest: fives,
	}
}

func faultResult(decision Decision, reason Reason, full bool) ParseResult {
	return ParseResult{Decision: decision, Reason: reason, Faults: []Reason{reason},
		FullSuiteRequired: full, EnforcementEffect: EnforcementNoEffect}
}

var passReference = ParseResult{Decision: DecisionPass, Reason: ReasonNone,
	Faults: []Reason{}, EnforcementEffect: EnforcementNoEffect}

func TestEnumsTagsAndFraming(t *testing.T) {
	decisions := [...]string{"PASS", "UNKNOWN", "FAIL_CLOSED"}
	reasons := [...]string{
		"NONE", "REQUIRED_INPUT_MISSING", "INVALID_UTF8", "BOM_FORBIDDEN",
		"INVALID_JSON", "TRAILING_VALUE", "DUPLICATE_KEY", "UNKNOWN_FIELD",
		"NULL_VALUE", "EMPTY_VALUE", "INVALID_SCHEMA", "INVALID_STABLE_ID",
		"INVALID_DIGEST", "BINDING_DIGEST_MISMATCH",
	}
	for i, want := range decisions {
		assert(t, uint8(Decision(i)) == uint8(i) && decisionSpelling(Decision(i)) == want, "decision")
	}
	for i, want := range reasons {
		assert(t, uint8(Reason(i)) == uint8(i) && reasonSpelling(Reason(i)) == want, "reason")
	}
	assert(t, EnforcementNoEffect == 0 && effectSpelling(EnforcementNoEffect) == "NO_EFFECT", "effect")
	for i, tag := range [...]byte{tagString, tagStableID, tagDigest, tagLegacyWorkID, tagEnum, tagBool, tagReasonList, tagU64} {
		assert(t, tag == byte(i+1), "tag")
	}
	assert(t, hex.EncodeToString(appendU64(nil, 0x0102030405060708)) == "0102030405060708", "big endian")
	assert(t, hex.EncodeToString(frame("d\x00", frameField{"x", tagU64, []byte{1, 2}})) == "0000000000000002640000000000000000010000000000000001780800000000000000020102", "frame")
}

func TestVectorsAndSelfExclusion(t *testing.T) {
	b := baseBinding()
	bd := bindingDigest(b)
	assert(t, len(bindingFrame(b)) == 772, "binding length")
	assert(t, bd == dBind, "binding digest")
	r := passReference
	rd := resultDigest(r)
	assert(t, len(resultFrame(r)) == 255, "result length")
	assert(t, rd == dResult, "result digest")
	b.BindingDigest, r.ResultDigest = bd, rd
	rp := replayDigest(b.BindingDigest, r)
	assert(t, len(replayFrame(b.BindingDigest, r)) == 252, "replay length")
	assert(t, rp == dReplay, "replay digest")
	b.BindingDigest = "sha256:mutated"
	assert(t, bindingDigest(b) == bd, "binding self digest")
	b.BindingDigest = bd
	r.ResultDigest = "sha256:mutated"
	assert(t, resultDigest(r) == rd, "result self digest")
	r.ResultDigest, r.ReplayDigest = rd, "sha256:mutated"
	assert(t, replayDigest(b.BindingDigest, r) == rp, "replay self digest")
	nilFaults := passReference
	nilFaults.Faults = nil
	assert(t, resultDigest(nilFaults) == rd, "nil faults")
}

func TestBindingMutationsAndFaultVectors(t *testing.T) {
	b := baseBinding()
	base := bindingDigest(b)
	mutations := []struct {
		name   string
		mutate func(*SafeWorkBinding)
	}{
		{"schema", func(v *SafeWorkBinding) { v.Schema += "-v2" }},
		{"task", func(v *SafeWorkBinding) { v.TaskID += "-v2" }},
		{"path", func(v *SafeWorkBinding) { v.PathID += "-v2" }},
		{"obligation", func(v *SafeWorkBinding) { v.ObligationID += "-v2" }},
		{"source", func(v *SafeWorkBinding) { v.SourceSnapshotDigest += "-v2" }},
		{"semantic", func(v *SafeWorkBinding) { v.SemanticSnapshotDigest += "-v2" }},
		{"policy", func(v *SafeWorkBinding) { v.PolicyDigest += "-v2" }},
		{"registry", func(v *SafeWorkBinding) { v.RegistryDigest += "-v2" }},
		{"toolchain", func(v *SafeWorkBinding) { v.ToolchainOptionsDigest += "-v2" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			changed := b
			mutation.mutate(&changed)
			assert(t, bindingDigest(changed) != base, "binding mutation")
		})
	}
	path := b
	path.PathID = "billing://path/pay-v2"
	registry := b
	registry.RegistryDigest = "sha256:9999999999999999999999999999999999999999999999999999999999999999"
	assert(t, bindingDigest(path) == dPath, "path vector")
	assert(t, bindingDigest(registry) == dReg, "registry vector")
	faults := []faultVector{
		{"missing", faultResult(DecisionUnknown, ReasonRequiredInputMissing, true), 306, dMissing},
		{"duplicate", faultResult(DecisionFailClosed, ReasonDuplicateKey, false), 292, dDuplicate},
		{"unknown", faultResult(DecisionFailClosed, ReasonUnknownField, false), 290, dUnknown},
		{"bom", faultResult(DecisionFailClosed, ReasonBOMForbidden, false), 292, dBOM},
		{"mismatch", faultResult(DecisionFailClosed, ReasonBindingDigestMismatch, false), 312, dMismatch},
	}
	for _, vector := range faults {
		t.Run(vector.name, func(t *testing.T) {
			assert(t, len(resultFrame(vector.result)) == vector.length, "fault length")
			assert(t, resultDigest(vector.result) == vector.digest, "fault digest")
			assert(t, vector.result.ReplayDigest == "", "fault replay")
		})
	}
}

func TestSurfaceOrderAndMetadata(t *testing.T) {
	typ := reflect.TypeOf(SafeWorkBinding{})
	fields := []struct{ name, tag string }{
		{"Schema", "schema"}, {"TaskID", "task_id"}, {"PathID", "path_id"},
		{"ObligationID", "obligation_id"}, {"SourceSnapshotDigest", "source_snapshot_digest"},
		{"SemanticSnapshotDigest", "semantic_snapshot_digest"}, {"PolicyDigest", "policy_digest"},
		{"RegistryDigest", "registry_digest"}, {"ToolchainOptionsDigest", "toolchain_options_digest"},
		{"BindingDigest", "binding_digest"},
	}
	assert(t, typ.NumField() == len(fields), "field count")
	for i, want := range fields {
		field := typ.Field(i)
		assert(t, field.Name == want.name, "field name")
		assert(t, field.Tag.Get("json") == want.tag, "json tag")
	}
	_, hasLegacy := typ.FieldByName("LegacyWorkID")
	assert(t, !hasLegacy && string(LegacyWorkID("raw")) == "raw", "legacy type")
	b := baseBinding()
	other := SafeWorkBinding{
		ToolchainOptionsDigest: b.ToolchainOptionsDigest, RegistryDigest: b.RegistryDigest,
		PolicyDigest: b.PolicyDigest, SemanticSnapshotDigest: b.SemanticSnapshotDigest,
		SourceSnapshotDigest: b.SourceSnapshotDigest, ObligationID: b.ObligationID,
		PathID: b.PathID, TaskID: b.TaskID, Schema: b.Schema,
	}
	assert(t, bindingDigest(b) == bindingDigest(other), "field order")
	metadata := struct{ name, expectedLabel string }{"first", "PASS"}
	before := resultDigest(passReference)
	metadata.name, metadata.expectedLabel = "second", "FAIL"
	assert(t, metadata.name != "first" && resultDigest(passReference) == before, "metadata")
}
