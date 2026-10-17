package toolchainformatfix

import (
	"bytes"
	"encoding/json"
	"testing"
)

func comparisonReportFixture() Report {
	report := Evaluate(Input{ExpectedHeadSHA: testHead})
	report.Cases[0].First.PeakRSSKiB = 17
	report.Cases[0].Replay.PeakRSSKiB = 23
	return sealComparisonFixture(report)
}

func sealComparisonFixture(report Report) Report {
	report.Cases = append([]CaseResult(nil), report.Cases...)
	for index := range report.Cases {
		report.Cases[index].EvidenceDigest = caseDigest(report.Cases[index])
	}
	report.Proofs = append([]Proof(nil), report.Proofs...)
	for index := range report.Proofs {
		if report.Proofs[index].Choice == "REGRESSION" {
			report.Proofs[index].EvidenceDigest = digestJSON(report.Cases)
		}
	}
	return seal(report)
}

func TestReplayReportComparisonPreservesRawResources(t *testing.T) {
	first := comparisonReportFixture()
	replay := sealComparisonFixture(first)
	replay.Cases[0].First.PeakRSSKiB += 100
	replay.Cases[0].Replay.PeakRSSKiB += 200
	replay = sealComparisonFixture(replay)
	before, err := json.Marshal([]Report{first, replay})
	if err != nil {
		t.Fatal(err)
	}
	if CompareReplayReports(first, replay, testHead) != nil ||
		CompareReplayReports(replay, first, testHead) != nil {
		t.Fatal("runner resource values changed otherwise equal report observations")
	}
	after, err := json.Marshal([]Report{first, replay})
	if err != nil || !bytes.Equal(before, after) ||
		first.Decision != DecisionClosed || replay.Decision != DecisionClosed {
		t.Fatal("comparison mutated raw records or authorized unresolved reports", err)
	}
}

func TestReplayReportComparisonRetainsSemanticDifferences(t *testing.T) {
	first := comparisonReportFixture()
	changes := []struct {
		name   string
		mutate func(*Report)
	}{
		{"arguments", func(r *Report) { r.Cases[0].First.Arguments = []string{"changed"} }},
		{"declared-arguments", func(r *Report) { r.Cases[0].Arguments = []string{"changed"} }},
		{"exit", func(r *Report) { r.Cases[0].First.ExitCode = 2 }},
		{"stdout", func(r *Report) { r.Cases[0].First.Stdout = "changed" }},
		{"stderr", func(r *Report) { r.Cases[0].First.Stderr = "changed" }},
		{"failure", func(r *Report) { r.Cases[0].First.Failure = "INVOCATION_DEADLINE" }},
		{"before", func(r *Report) { r.Cases[0].First.TreeBeforeDigest = digestJSON("before") }},
		{"after", func(r *Report) { r.Cases[0].First.TreeAfterDigest = digestJSON("after") }},
		{"writes", func(r *Report) { r.Cases[0].First.RepositoryWrites = 1 }},
		{"reason", func(r *Report) { r.Cases[0].Reason = "changed" }},
		{"definition", func(r *Report) { r.Cases[0].Definition.ExpectedExit = 2 }},
		{"replay-output", func(r *Report) { r.Cases[0].Replay.Stdout = "changed" }},
		{"toolchain", func(r *Report) { r.Source.GoVersion = "changed" }},
		{"executable", func(r *Report) { r.Source.ExecutableDigest = digestJSON("changed") }},
		{"summary", func(r *Report) { r.Summary.Invocations++ }},
		{"proof-operation", func(r *Report) { r.Proofs[0].MetaOperation = "changed" }},
		{"coherence-digest", func(r *Report) { r.Proofs[1].EvidenceDigest = digestJSON("changed") }},
		{"case-proof-status", func(r *Report) { r.Proofs[2].Passed = !r.Proofs[2].Passed }},
		{"case-status", func(r *Report) { r.Cases[0].Status = "SKIPPED" }},
		{"resolution", func(r *Report) { r.Resolution = ResolutionExact }},
		{"completed-count", func(r *Report) { r.Summary.Executed++ }},
		{"source-availability", func(r *Report) { r.Source.ObservationKnown = true }},
	}
	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			replay := sealComparisonFixture(first)
			change.mutate(&replay)
			replay = sealComparisonFixture(replay)
			if err := Validate(replay, testHead); err != nil {
				t.Fatal("difference fixture must retain a valid report envelope", err)
			}
			if CompareReplayReports(first, replay, testHead) == nil {
				t.Fatal("semantic difference disappeared from comparison", change.name)
			}
		})
	}
}

func TestReplayReportComparisonRejectsInvalidBindings(t *testing.T) {
	first := comparisonReportFixture()
	for _, name := range []string{
		"report-digest", "case-digest", "case-shape", "schema", "caller-head",
		"case-proof-digest", "case-proof-role", "case-proof-duplicate",
		"decision", "case-proof-missing", "source-head",
	} {
		t.Run(name, func(t *testing.T) {
			replay := sealComparisonFixture(first)
			head := testHead
			switch name {
			case "report-digest":
				replay.ReportDigest = ""
			case "case-digest":
				replay.Cases[0].EvidenceDigest = ""
				replay = seal(replay)
			case "case-shape":
				replay.Cases = replay.Cases[1:]
				replay = seal(replay)
			case "schema":
				replay.Schema = "unknown"
				replay = seal(replay)
			case "caller-head":
				head = "1111111111111111111111111111111111111111"
			case "case-proof-digest":
				replay.Proofs[2].EvidenceDigest = ""
				replay = seal(replay)
			case "case-proof-role":
				replay.Proofs[2].Choice = "UNKNOWN"
				replay = seal(replay)
			case "case-proof-duplicate":
				replay.Proofs[0] = replay.Proofs[2]
				replay = seal(replay)
			case "decision":
				replay.Decision = DecisionPass
				replay = seal(replay)
			case "case-proof-missing":
				replay.Proofs = replay.Proofs[:2]
				replay = seal(replay)
			case "source-head":
				replay.Source.ExpectedHeadSHA = "1111111111111111111111111111111111111111"
				replay = seal(replay)
			}
			if CompareReplayReports(first, replay, head) == nil ||
				CompareReplayReports(replay, first, head) == nil {
				t.Fatal("invalid raw report binding survived comparison", name)
			}
		})
	}
}
