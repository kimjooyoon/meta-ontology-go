package integrationprogress

import "fmt"

func completeObservation() Observation {
	value := Observation{Schema: ObservationSchema, Repository: Repository,
		ObserverHeadSHA: fmt.Sprintf("%040x", 999), ObservedAt: "2026-08-28T13:00:00Z",
		CohortID: CohortID, QueueSnapshot: QueueSnapshot{ObservationStatus: "OBSERVED", QueuedRuns: 3, InProgressRuns: 1},
		PullRequests: make([]PullObservation, 0, len(PullNumbers()))}
	for _, number := range PullNumbers() {
		head := fmt.Sprintf("%040x", number)
		value.PullRequests = append(value.PullRequests, PullObservation{
			Number: number, ObservationStatus: "OBSERVED", State: "closed", HeadSHA: head,
			CreatedAt: "2026-08-01T12:00:00Z", ClosedAt: "2026-08-01T12:04:00Z",
			MergedAt: "2026-08-01T12:04:00Z", RunsTotal: 1, RunsConsumed: 1,
			RunSelection: "LATEST_TERMINAL_BEFORE_MERGE", EligibleRuns: 1,
			AuthoritativeRun: &RunObservation{ID: int64(number), Name: WorkflowName, HeadSHA: head,
				Status: "completed", Conclusion: "success", CreatedAt: "2026-08-01T12:00:00Z",
				StartedAt: "2026-08-01T12:01:00Z", UpdatedAt: "2026-08-01T12:03:00Z",
				ArtifactsTotal: 1, ArtifactsConsumed: 1, ArtifactMatches: 1,
				Artifact: &ArtifactObservation{ID: int64(number), Name: ArtifactPrefix + head,
					HeadSHA: head, CreatedAt: "2026-08-01T12:02:00Z"}},
		})
	}
	return value
}

func fixturePull(value *Observation, number int) *PullObservation {
	for index := range value.PullRequests {
		if value.PullRequests[index].Number == number {
			return &value.PullRequests[index]
		}
	}
	return nil
}
