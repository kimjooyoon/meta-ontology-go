package safeworkbinding

import (
	"testing"
)

const (
	digestBase          = "sha256:dc6dbe157ede5924b61676bfdcd4151cd6f73a51b7eefda674cca3d6d169a5cb"
	digestPass          = "sha256:b076bd08e4b82b4b2aeb78f3f8a3e12931ae8f8a01f734cb2a97bca15753a342"
	digestReplay        = "sha256:4564f79ab93fadeb3221f837e38dfdce43e1c653a89086283aaceb09d923080d"
	digestPath          = "sha256:c262c10f1652d0d3fd5fa605166d565a44261b834ffcffe876f2bb785f4bcd51"
	digestRegistry      = "sha256:23477b01e3d477d6ba5702adca8ad77a2cd643748b29e20ac4a7fda0b020673e"
	digestMissing       = "sha256:3b4256d3ab6d60db4caa093cc8306819c9a6fb9fb987927fab8bff841a0690c8"
	digestDuplicate     = "sha256:bac3126c6e354dfb38abf69508605e50226f186d25b7381f8eb3215f33e7fda0"
	digestExpectedLabel = "sha256:1ada140e4d914ab1ab15570deb11e423a7e454b81c2bbf84512454b642ecfa02"
	digestMismatch      = "sha256:c61d5823d85939c5f7645d6f2fe2049fe724bfec3fc12c011e9532ccc41c6b10"
	digestBOMMixed      = "sha256:49d26986dde23834ba79d1064aa577860dfe47a79bff2ab698d385660db1ac8c"
)

func repeatedDigest(digit byte) Digest {
	value := make([]byte, 64)
	for i := range value {
		value[i] = digit
	}
	return Digest("sha256:" + string(value))
}
func baseBindingForDigest() SafeWorkBinding {
	return SafeWorkBinding{
		Schema:                 SafeWorkBindingSchemaV1,
		TaskID:                 "billing://task/pay",
		PathID:                 "billing://path/pay",
		ObligationID:           "billing://obligation/pay",
		SourceSnapshotDigest:   repeatedDigest('1'),
		SemanticSnapshotDigest: repeatedDigest('2'),
		PolicyDigest:           repeatedDigest('3'),
		RegistryDigest:         repeatedDigest('4'),
		ToolchainOptionsDigest: repeatedDigest('5'),
	}
}
func passResultForDigest() ParseResult {
	return ParseResult{Decision: DecisionPass, Reason: ReasonNone, Faults: []Reason{},
		EnforcementEffect: EnforcementEffectNoEffect}
}
func fixtureResult(decision Decision, reason Reason, fullSuite bool) ParseResult {
	return ParseResult{Decision: decision, Reason: reason, Faults: []Reason{reason},
		FullSuiteRequired: fullSuite, EnforcementEffect: EnforcementEffectNoEffect}
}
func checkResultDigest(t *testing.T, result ParseResult, want Digest) {
	t.Helper()
	frame, frameOK := resultFrame(result)
	check(t, frameOK, "result frame")
	check(t, len(frame) > 0, "result frame bytes")
	digest, digestOK := resultDigest(result)
	check(t, digestOK, "result digest")
	check(t, digest == want, "result vector")
}
