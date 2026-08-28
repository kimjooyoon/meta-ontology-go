package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func loadSubjects(name, expectedSHA string) ([]inputSubject, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var evidence projectionEvidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		return nil, err
	}
	if evidence.Schema != "gooo.repository-projection-evidence.v1" {
		return nil, fmt.Errorf("unsupported projection evidence %q", evidence.Schema)
	}
	if evidence.SourceSHA != expectedSHA {
		return nil, fmt.Errorf("evidence SHA %s does not match %s",
			evidence.SourceSHA, expectedSHA)
	}
	selected := make([]inputSubject, 0, len(evidence.Subjects))
	seen := make(map[string]bool)
	for _, subject := range evidence.Subjects {
		if subject.Indicator != "source.line-cap-debt" {
			continue
		}
		if seen[subject.Logical] {
			return nil, fmt.Errorf("duplicate subject %s", subject.Logical)
		}
		if subject.Limit != 75 || subject.Value <= subject.Limit {
			return nil, fmt.Errorf("invalid line-cap evidence for %s", subject.Logical)
		}
		seen[subject.Logical] = true
		selected = append(selected, subject)
	}
	return selected, nil
}

func loadMetricSubjects(name, expectedSHA string) ([]inputSubject, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var report linecaps.LineMetricsReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	if report.CommitSHA != expectedSHA {
		return nil, fmt.Errorf("metrics SHA %s does not match %s", report.CommitSHA, expectedSHA)
	}
	selected := make([]inputSubject, 0)
	seen := make(map[string]bool)
	for _, indicator := range report.Meta.Indicators {
		if indicator.MetricID != sourcepolicy.DimensionGoFileLines ||
			indicator.Applicability == sourcepolicy.ApplicabilityNotApplicable ||
			indicator.Satisfied || !indicator.Blocking ||
			indicator.Operation != sourcepolicy.OperationSplitGo {
			continue
		}
		if indicator.Subject == "" || seen[indicator.Subject] {
			if indicator.Subject == "" {
				return nil, fmt.Errorf("metrics contain an empty source subject")
			}
			return nil, fmt.Errorf("duplicate metric subject %s", indicator.Subject)
		}
		seen[indicator.Subject] = true
		selected = append(selected, inputSubject{
			Indicator: string(indicator.MetricID), Logical: indicator.Subject,
			Value: indicator.Value, Limit: indicator.Limit,
			Consumer: string(indicator.Consumer), Operation: string(indicator.Operation),
		})
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Logical < selected[j].Logical })
	return selected, nil
}

func writePackagePartitionReceipt(name, sha string, recipe packagePartitionRecipe, writes map[string]bool, fixed bool) error {
	paths := make([]string, 0, len(writes))
	for path := range writes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	decision, reason := "PASS", "GO_PACKAGE_PARTITION_APPLIED"
	moved, created, rewrites := len(recipe.Moves), len(recipe.Creates), len(recipe.Rewrites)+len(recipe.Ranges)
	if fixed {
		decision, reason = "FIXED_POINT", "GO_PACKAGE_PARTITION_ALREADY_APPLIED"
		moved, created, rewrites = 0, 0, 0
	}
	receipt := map[string]any{"schema": "gooo.go-package-partition-receipt.v1", "decision": decision, "resolution": "EXACT", "reason": reason, "source_sha": sha, "subject": recipe.Subject, "meta_operation": "partition-go-package", "moved_files": moved, "created_files": created, "rewrites": rewrites, "write_set": paths, "expected_shape": recipe.ExpectedShape, "effects": map[string]any{"repository_writes": 0, "mutation_authority": false, "disposable_writes": len(paths)}}
	sealed, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(sealed)
	receipt["digest"] = "sha256:" + hex.EncodeToString(digest[:])
	payload, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(payload, '\n'), 0o644)
}
