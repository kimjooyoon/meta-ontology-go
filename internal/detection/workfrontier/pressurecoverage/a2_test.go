package pressurecoverage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

const (
	a2Snapshot   = "sha256:875f886774b6c7127f6b59ee5cb5facaf5825a36f708900ba63235d7db2e9b8f"
	a2Policy     = "sha256:a4d888b25b683488a6751d9dcc487043002be60441122fe8a87afc54b809fa49"
	a2Registry   = "sha256:e0d1d311e52cc85a3ff82cf7b49b299fa7dfa7bcf876eb0995e838188dd24c57"
	a2Toolchain  = "sha256:9f2f29d60c221e56bf389b6721e08875db5f7d6b14b30d9c25b8fc73e6908cb2"
	a2InputHash  = "sha256:c6dbf237e8c44a55c79836153a6880fabdfaaa3d8e872dc07e99e337cd2e4fd3"
	a2ResultHash = "sha256:620f8ba049427fcc6dff30f4592ee10acbdaf931d5075056b9e35128bb41afa5"
	a2ReplayHash = "sha256:de2a612f99485c5711b709dc804e4d289c88a1d283b9e55427691b8ce2a29a78"
)

type a2PrecedenceCase struct {
	name  string
	bind  bool
	edit  func(*Input)
	wantD Decision
	wantR Reason
}

var a2PrecedenceCases = []a2PrecedenceCase{
	{
		name:  "malformed A1 input",
		edit:  func(input *Input) { input.RequiredPressureIDs[0] = "p x" },
		wantD: DecisionFailClosed,
		wantR: ReasonInvalidInput,
	},
	{
		name:  "blank binding",
		edit:  func(input *Input) { input.PolicyDigest = "" },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "stale binding",
		edit:  func(input *Input) { input.PolicyDigest = "stale" },
		wantD: DecisionUnknown,
		wantR: ReasonSnapshotMismatch,
	},
	{
		name:  "zero K",
		bind:  true,
		edit:  func(input *Input) { input.RequestedK = 0 },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "zero minimum",
		bind:  true,
		edit:  func(input *Input) { input.MinimumIndependent = 0 },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "K below floor",
		bind:  true,
		edit:  func(input *Input) { input.RequestedK = 1 },
		wantD: DecisionFailClosed,
		wantR: ReasonPolicyFloorViolation,
	},
	{
		name:  "minimum below floor",
		bind:  true,
		edit:  func(input *Input) { input.MinimumIndependent = 1 },
		wantD: DecisionFailClosed,
		wantR: ReasonPolicyFloorViolation,
	},
	{
		name:  "minimum above K",
		bind:  true,
		edit:  func(input *Input) { input.MinimumIndependent = 3 },
		wantD: DecisionFailClosed,
		wantR: ReasonPolicyFloorViolation,
	},
	{
		name:  "empty required",
		bind:  true,
		edit:  func(input *Input) { input.RequiredPressureIDs = nil },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "missing record",
		bind:  true,
		edit:  func(input *Input) { input.PressureRecords = input.PressureRecords[1:] },
		wantD: DecisionUnknown,
		wantR: ReasonRequiredInputMissing,
	},
	{
		name:  "blank group",
		bind:  true,
		edit:  func(input *Input) { input.PressureRecords[0].IndependenceGroupID = "" },
		wantD: DecisionUnknown,
		wantR: ReasonApplicabilityOrGroupUnproven,
	},
	{
		name:  "blank applicability",
		bind:  true,
		edit:  func(input *Input) { input.PressureRecords[0].ApplicabilityRuleID = "" },
		wantD: DecisionUnknown,
		wantR: ReasonApplicabilityOrGroupUnproven,
	},
	{
		name: "same group",
		bind: true,
		edit: func(input *Input) {
			for index := range input.PressureRecords {
				input.PressureRecords[index].IndependenceGroupID = "group-a"
			}
		},
		wantD: DecisionUnknown,
		wantR: ReasonIndependentGroupShortfall,
	},
}

func TestEvaluatePositiveResult(t *testing.T) {
	got := Evaluate(a2Input())
	if got.Decision != DecisionPass || got.Reason != ReasonNone ||
		got.RequiredPressureCount != 3 || got.DistinctGroupCount != 3 {
		t.Fatalf("result = %#v", got)
	}
	if got.InputDigest != a2InputHash || got.ResultDigest != a2ResultHash || got.ReplayDigest != a2ReplayHash {
		t.Fatalf("digests = %#v", got)
	}
	if !reflect.DeepEqual(got.RequiredPressureIDs, []string{"p-a", "p-b", "p-c"}) ||
		!reflect.DeepEqual(got.RequiredGroupIDs, []string{"group-a", "group-b", "group-c"}) ||
		!reflect.DeepEqual(got.MissingPressureIDs, []string{}) {
		t.Fatalf("canonical sets = %#v", got)
	}
}

func TestEvaluatePrecedence(t *testing.T) {
	for _, test := range a2PrecedenceCases {
		t.Run(test.name, func(t *testing.T) {
			input := a2Input()
			test.edit(&input)
			if test.bind {
				testBind(&input)
			}
			got := Evaluate(input)
			if got.Decision != test.wantD || got.Reason != test.wantR {
				t.Fatalf("result = %#v", got)
			}
		})
	}
}

func TestEvaluateK21IsInputDriven(t *testing.T) {
	input := a2Input()
	input.RequestedK = 21
	for number := 4; number <= 21; number++ {
		id := fmt.Sprintf("p-%02d", number)
		input.PressureRecords = append(input.PressureRecords,
			PressureRecord{id, "category-" + id, "group-" + id, "rule-1"})
		input.RequiredPressureIDs = append(input.RequiredPressureIDs, id)
	}
	testBind(&input)
	got := Evaluate(input)
	if got.Decision != DecisionPass || got.RequiredPressureCount != 21 || got.DistinctGroupCount != 21 {
		t.Fatalf("K=21 result = %#v", got)
	}
}

func TestEvaluatePermutationReplay(t *testing.T) {
	base := Evaluate(a2Input())
	input := a2Input()
	input.RequiredPressureIDs[0], input.RequiredPressureIDs[2] = input.RequiredPressureIDs[2], input.RequiredPressureIDs[0]
	input.PressureRecords[0], input.PressureRecords[2] = input.PressureRecords[2], input.PressureRecords[0]
	got := Evaluate(input)
	if !reflect.DeepEqual(got, base) || !bytes.Equal(resultBytes(got), resultBytes(base)) {
		t.Fatalf("permutation changed result: %#v != %#v", got, base)
	}
}

func resultBytes(result Result) []byte {
	data, _ := json.Marshal(result)
	return data
}

func a2Input() Input {
	return Input{
		Schema:                  SchemaVersion,
		AuthoritySnapshotDigest: a2Snapshot,
		PolicyDigest:            a2Policy,
		RegistryDigest:          a2Registry,
		ToolchainOptionsDigest:  a2Toolchain,
		RequestedK:              2,
		MinimumIndependent:      2,
		PressureRecords: []PressureRecord{
			{"p-c", "category-c", "group-c", "rule-1"},
			{"p-a", "category-a", "group-a", "rule-1"},
			{"p-b", "category-b", "group-b", "rule-1"},
		},
		RequiredPressureIDs: []string{"p-c", "p-a", "p-b"},
	}
}

// testBind independently rebinds mutated fixtures with raw SHA-256 construction.
func testBind(input *Input) {
	unsigned := *input
	unsigned.AuthoritySnapshotDigest = ""
	unsigned.PolicyDigest = ""
	unsigned.RegistryDigest = ""
	unsigned.ToolchainOptionsDigest = ""
	unsigned.PressureRecords = append([]PressureRecord(nil), unsigned.PressureRecords...)
	unsigned.RequiredPressureIDs = append([]string(nil), unsigned.RequiredPressureIDs...)
	sort.Slice(unsigned.PressureRecords, func(left, right int) bool {
		return unsigned.PressureRecords[left].PressureID < unsigned.PressureRecords[right].PressureID
	})
	sort.Strings(unsigned.RequiredPressureIDs)
	data, _ := json.Marshal(unsigned)
	inputDigest := testDigest(data)
	input.AuthoritySnapshotDigest = testRoleDigest("authority-snapshot", inputDigest)
	input.PolicyDigest = testRoleDigest("policy", inputDigest)
	input.RegistryDigest = testRoleDigest("registry", inputDigest)
	input.ToolchainOptionsDigest = testRoleDigest("toolchain-options", inputDigest)
}

func testRoleDigest(role, inputDigest string) string {
	return testDigest([]byte(role + "\x00" + inputDigest))
}

func testDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
