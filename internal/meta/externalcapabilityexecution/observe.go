package externalcapabilityexecution

import (
	"encoding/json"
	"fmt"
	"os"
)

func Observe(options ObserverOptions, parentJSON []byte) Observation {
	observation := Observation{
		Schema: ObservationSchema, SubjectSHA: options.SubjectSHA, Available: true,
		UnknownEvents: []string{},
	}
	if err := json.Unmarshal(parentJSON, &observation.Parent); err != nil {
		return unavailable(observation, fmt.Errorf("decode parent report: %w", err))
	}
	reference, err := observeReference(options.ExternalRoot)
	if err != nil {
		return unavailable(observation, err)
	}
	observation.Reference = reference
	sourceBefore, err := dirtyPaths(options.SourceRoot)
	if err != nil {
		return unavailable(observation, err)
	}
	externalBefore, err := dirtyPaths(options.ExternalRoot)
	if err != nil {
		return unavailable(observation, err)
	}
	workspace, err := os.MkdirTemp("", "gooo-external-capability-")
	if err != nil {
		return unavailable(observation, err)
	}
	defer os.RemoveAll(workspace)
	observation.Runs, observation.ExternalExecutions, err = executeCapabilities(workspace, options.ExternalRoot)
	sourceAfter, sourceErr := dirtyPaths(options.SourceRoot)
	externalAfter, externalErr := dirtyPaths(options.ExternalRoot)
	observation.RepositoryWrites = maximum(sourceBefore, sourceAfter)
	observation.ExternalRepositoryWrites = maximum(externalBefore, externalAfter)
	if err != nil || sourceErr != nil || externalErr != nil {
		return unavailable(observation, fmt.Errorf("execution=%v source=%v external=%v", err, sourceErr, externalErr))
	}
	observation.ReplayExact = len(observation.Runs) == 2 &&
		observation.Runs[0].NormalizedSHA256 == observation.Runs[1].NormalizedSHA256
	return sealObservation(observation)
}

func unavailable(observation Observation, err error) Observation {
	observation.Available = false
	observation.UnknownEvents = append(observation.UnknownEvents, err.Error())
	return sealObservation(observation)
}

func maximum(left, right int) int {
	if left > right {
		return left
	}
	return right
}
