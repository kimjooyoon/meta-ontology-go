package toolchainformatfix

import (
	"errors"
	"testing"

	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

func TestPartialInvocationReportComparisonRetainsUnresolvedState(t *testing.T) {
	for _, mode := range []string{"error", "marker"} {
		t.Run(mode, func(t *testing.T) {
			first := partialReplayReport(t, mode, 17)
			replay := partialReplayReport(t, mode, 23)
			before := digestJSON([]Report{first, replay})
			if err := CompareReplayReports(first, replay, testHead); err != nil {
				t.Fatal("otherwise equal partial observations differed", err)
			}
			if before != digestJSON([]Report{first, replay}) {
				t.Fatal("comparison rewrote partial observations or raw resource values")
			}
			for _, report := range []Report{first, replay} {
				if report.Decision != DecisionClosed || report.Resolution != ResolutionLower ||
					report.Summary.Unresolved != FixedTotal || report.Summary.Satisfied != 0 ||
					report.RepositoryWrites != 3*FixedTotal || report.MutationAuthorized {
					t.Fatal("equal partial reports acquired successful execution authority", report)
				}
			}
		})
	}
}

func partialReplayReport(t *testing.T, mode string, rss int64) Report {
	t.Helper()
	calls := 0
	executor := partialObservationExecutor{invoke: func(arguments []string) (cliruntime.Observation, error) {
		calls++
		observation := formatFixPartialObservation(arguments)
		observation.PeakRSSKiB, observation.RepositoryWrites = rss, 3
		observation.Stderr, observation.TreeAfterDigest = "retained partial stderr", digestJSON("changed tree")
		if mode == "error" {
			return observation, errors.New("controlled partial observation")
		}
		observation.Failure = "INVOCATION_RESOURCE_UNKNOWN"
		return observation, nil
	}}
	report := Report{Schema: ReportSchema, Source: Source{ExpectedHeadSHA: testHead}}
	for _, definition := range expectedRegistry().Cases {
		result := executeCase(executor, definition)
		if result.Status != "UNRESOLVED" || result.ReplayMatched || len(result.Replay.Arguments) != 0 {
			t.Fatal("partial invocation was accepted or replayed", result)
		}
		report.Cases = append(report.Cases, result)
	}
	if calls != FixedTotal {
		t.Fatal("incomplete first invocations were replayed", calls)
	}
	return finish(report, cycleEvidence{}, 0, "FORMAT_FIX_OBSERVATION_UNKNOWN")
}
