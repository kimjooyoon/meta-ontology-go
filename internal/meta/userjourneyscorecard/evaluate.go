package userjourneyscorecard

import (
	"bytes"
	"encoding/json"
	"fmt"

	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
)

type inspection struct {
	contract Contract
	profile  Profile
	upstream metacli.Report
	stats    []JourneyStats
	upstreamPassed, binaryBound, sourceBound bool
	journeysPassed, envelopesPassed          int
	samplesObserved, outputReplays            int
	metaBindings, unknowns                    int
	wallViolations, rssViolations             int
	binaryViolations, repositoryWrites        int
	lowerResolution                           bool
}

func Evaluate(root, executable, head string, contractRaw, upstreamRaw, profileRaw []byte) (Report, error) {
	first, err := evaluateOnce(root, executable, head, contractRaw, upstreamRaw, profileRaw)
	if err != nil {
		return Report{}, err
	}
	second, err := evaluateOnce(root, executable, head, contractRaw, upstreamRaw, profileRaw)
	if err != nil {
		return Report{}, err
	}
	if !bytes.Equal(mustJSON(first), mustJSON(second)) {
		return Report{}, fmt.Errorf("scorecard reducer replay diverged")
	}
	first.ReducerReplayVerified = true
	seal(&first)
	return first, nil
}

func evaluateOnce(root, executable, head string, contractRaw, upstreamRaw, profileRaw []byte) (Report, error) {
	s := inspection{}
	if err := json.Unmarshal(contractRaw, &s.contract); err != nil {
		return Report{}, fmt.Errorf("decode contract: %w", err)
	}
	if err := json.Unmarshal(upstreamRaw, &s.upstream); err != nil {
		return Report{}, fmt.Errorf("decode upstream: %w", err)
	}
	if err := json.Unmarshal(profileRaw, &s.profile); err != nil {
		return Report{}, fmt.Errorf("decode profile: %w", err)
	}
	if err := s.inspectEvidence(root, executable, head); err != nil {
		return Report{}, err
	}
	s.reduceSamples()
	return s.buildReport(head, contractRaw, upstreamRaw, profileRaw), nil
}
