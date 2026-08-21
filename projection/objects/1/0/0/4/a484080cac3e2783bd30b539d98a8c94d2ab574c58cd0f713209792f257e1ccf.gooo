package pressureshadow

import (
	"reflect"
	"testing"
)

const (
	s1b1Snapshot    = "sha256:875f886774b6c7127f6b59ee5cb5facaf5825a36f708900ba63235d7db2e9b8f"
	s1b1Policy      = "sha256:a4d888b25b683488a6751d9dcc487043002be60441122fe8a87afc54b809fa49"
	s1b1Registry    = "sha256:e0d1d311e52cc85a3ff82cf7b49b299fa7dfa7bcf876eb0995e838188dd24c57"
	s1b1Toolchain   = "sha256:9f2f29d60c221e56bf389b6721e08875db5f7d6b14b30d9c25b8fc73e6908cb2"
	s1b1A2Input     = "sha256:c6dbf237e8c44a55c79836153a6880fabdfaaa3d8e872dc07e99e337cd2e4fd3"
	s1b1A2Result    = "sha256:620f8ba049427fcc6dff30f4592ee10acbdaf931d5075056b9e35128bb41afa5"
	s1b1A2Replay    = "sha256:de2a612f99485c5711b709dc804e4d289c88a1d283b9e55427691b8ce2a29a78"
	s1b1InputDigest = "sha256:fb0fc08308285457a13d6f100455f2d02e0a90faae554c2803cfdd730b9df2d9"
	s1b1Result      = "sha256:f6abb3a8c03de565f3e7e86642df3c5892209ce94431e7189b0d4ddce11dfcec"
	s1b1Replay      = "sha256:866b49b942fc917714c9a48d9eb6c6aa3c96390c4599196bb38bc9254b05daa3"
)

func TestS1B1PositiveAndPermutation(t *testing.T) {
	input := s1b1Input()
	got := ValidateS1B1(input)
	if got.Decision != DecisionPass || got.Reason != ReasonNone ||
		got.InputDigest != s1b1InputDigest || got.ResultDigest != s1b1Result || got.ReplayDigest != s1b1Replay ||
		!sameB1Values(got.PressureCoveragePassPathIDs, []string{"path/a", "path/b", "path/c"}) {
		t.Fatalf("positive result = %#v", got)
	}
	if got.A2Observations[0].Result.InputDigest != s1b1A2Input ||
		got.A2Observations[0].Result.ResultDigest != s1b1A2Result ||
		got.A2Observations[0].Result.ReplayDigest != s1b1A2Replay {
		t.Fatalf("path/a A2 result = %#v", got.A2Observations[0])
	}
	input.Selector.Paths[0], input.Selector.Paths[2] = input.Selector.Paths[2], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[2] = input.PathCoverage[2], input.PathCoverage[0]
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.RequiredPressureIDs[0], coverage.RequiredPressureIDs[2] =
			coverage.RequiredPressureIDs[2], coverage.RequiredPressureIDs[0]
		coverage.PressureRecords[0], coverage.PressureRecords[2] = coverage.PressureRecords[2], coverage.PressureRecords[0]
	}
	if replay := ValidateS1B1(input); !reflect.DeepEqual(replay, got) {
		t.Fatalf("permutation changed result: %#v", replay)
	}
}

var s1b1Unknown = ReasonPressureCoverageUnknown
var s1b1Fail = ReasonPressureCoverageFailClosed
