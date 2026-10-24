package sourcesplit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/exactsha"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

// NewObserver binds one metrics payload to the repository HEAD. Each candidate
// then uses the same renderer and topology checks as the writing command.
func NewObserver(root, head string, raw []byte) (generation.PrerequisiteObserver, error) {
	if err := exactsha.Verify(root, head); err != nil {
		return nil, err
	}
	var report metricevidence.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, err
	}
	if report.CommitSHA != head || report.Meta.Policy.MaxFileLines <= 0 || report.Meta.Policy.MaxDirectEntries <= 0 {
		return nil, fmt.Errorf("split prerequisite metrics revision or policy is missing")
	}
	metricsDigest := sourceDigest(raw)
	return func(action generation.Action) generation.CandidatePrerequisite {
		if generation.NativeExecutionRoute(action.Operation) == "" {
			return generation.NewCandidatePrerequisite(action, head, "", metricsDigest, "EXECUTION_CONSUMER_UNAVAILABLE")
		}
		digest, reason := observe(root, action.Subject, report)
		return generation.NewCandidatePrerequisite(action, head, digest, metricsDigest, reason)
	}, nil
}

func observe(root, subject string, report metricevidence.Report) (string, string) {
	target, err := secureSourcePath(root, subject)
	if err != nil {
		return "", "SOURCE_PLANNING_UNAVAILABLE"
	}
	before, err := os.ReadFile(target)
	if err != nil {
		return "", "SOURCE_PLANNING_UNAVAILABLE"
	}
	digest := sourceDigest(before)
	plan, planErr := PlanSource(root, subject, report.Meta.Policy.MaxFileLines)
	after, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(before, after) {
		return digest, "SOURCE_CHANGED_DURING_PLANNING"
	}
	if planErr != nil {
		if blocker, ok := errors.AsType[*Blocker](planErr); ok {
			return digest, blocker.Reason
		}
		return digest, "SOURCE_PLANNING_UNAVAILABLE"
	}
	return digest, observeTopology(report, plan)
}

func observeTopology(report metricevidence.Report, plan Plan) string {
	known := 0
	for _, directory := range report.Directories {
		if directory.Path == plan.Directory {
			known++
		}
	}
	if known != 1 {
		return "DIRECTORY_OBSERVATION_MISSING"
	}
	// Logical directory counts cannot discharge a physical successor projection.
	if report.StorageRoot != "" || len(report.StorageDirectories) != 0 {
		return "SUCCESSOR_PROJECTION_REQUIRED"
	}
	if err := ValidateTopology(report, plan); err != nil {
		if blocker, ok := errors.AsType[*Blocker](err); ok {
			return blocker.Reason
		}
		return "DIRECTORY_OBSERVATION_MISSING"
	}
	return "SPLIT_PLAN_READY"
}

func sourceDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
