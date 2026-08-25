package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	directIndicator        = "storage.direct-entry"
	directObservedIndicator = "storage.direct-entry-observed"
	directUnboundIndicator = "storage.direct-entry-unclassified"
	bootstrapDirectIndicator = "storage.bootstrap-direct-entry"
	mixedIndicator         = "storage.mixed-kind"
	maxStoredDirectEntries = 10
	maxStoredKinds         = 1
	topologyProof          = "axiomatic-foundation"
)

func topologyFailures(root string) (topologyEvidence, error) {
	result := topologyEvidence{}
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || !entry.IsDir() || filepath.Clean(name) == filepath.Clean(root) {
			return walkErr
		}
		children, err := os.ReadDir(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		physical := filepath.ToSlash(relative)
		if len(children) > maxStoredDirectEntries {
			result.ObservedDirect++
			item := subject{
				Indicator: directIndicator, Physical: physical, Value: len(children), Limit: maxStoredDirectEntries,
				Consumer: "radix-sharder", Operation: "split-object-bucket",
				Applicability: "APPLICABLE", ApplicabilityReason: "GENERIC_PHYSICAL_DIRECTORY",
			}
			if workflowDiscoveryRoot(physical, children) {
				result.ExemptDirect++
				item.Indicator = bootstrapDirectIndicator
				item.Consumer = "github-actions"
				item.Operation = "preserve-workflow-discovery"
				item.Applicability = "NOT_APPLICABLE"
				item.ApplicabilityReason = "GITHUB_WORKFLOW_DISCOVERY_ROOT"
			} else {
				result.Direct++
			}
			result.Subjects = append(result.Subjects, item)
		}
		hasDirectory, hasFile := false, false
		for _, child := range children {
			hasDirectory = hasDirectory || child.IsDir()
			hasFile = hasFile || !child.IsDir()
		}
		if hasDirectory && hasFile {
			result.Mixed++
			result.Subjects = append(result.Subjects, subject{
				Indicator: mixedIndicator, Physical: physical, Value: 2, Limit: maxStoredKinds,
				Consumer: "radix-sharder", Operation: "separate-branch-leaf",
			})
		}
		return nil
	})
	sort.Slice(result.Subjects, func(i, j int) bool {
		if result.Subjects[i].Indicator == result.Subjects[j].Indicator {
			return result.Subjects[i].Physical < result.Subjects[j].Physical
		}
		return result.Subjects[i].Indicator < result.Subjects[j].Indicator
	})
	return result, err
}

func workflowDiscoveryRoot(physical string, children []os.DirEntry) bool {
	if physical != ".github/workflows" || len(children) == 0 {
		return false
	}
	for _, child := range children {
		extension := strings.ToLower(filepath.Ext(child.Name()))
		if child.IsDir() || child.Type()&os.ModeSymlink != 0 || (extension != ".yml" && extension != ".yaml") {
			return false
		}
	}
	return true
}

func buildEvidence(sha string, model manifest, objects, loss int, topology topologyEvidence) evidence {
	unbound, lineDebt := 0, 0
	subjects := append([]subject(nil), topology.Subjects...)
	for _, entry := range model.Entries {
		if entry.ObjectSHA == "" || entry.Backing == "" {
			unbound++
		}
		if entry.Language != "" && entry.Lines > 75 {
			lineDebt++
			subjects = append(subjects, subject{
				Indicator: "source.line-cap-debt", Logical: entry.Logical,
				Value: entry.Lines, Limit: 75, Consumer: "logical-source-splitter",
				Operation: "split-before-storage",
			})
		}
	}
	proof := topologyProof
	unclassifiedDirect := topology.ObservedDirect - topology.Direct - topology.ExemptDirect
	return evidence{
		Schema: "gooo.repository-projection-evidence.v1", SourceSHA: sha,
		TrackedFiles: len(model.Entries), Objects: objects, Subjects: subjects,
		Indicators: []indicator{
			{ID: "projection.roundtrip-loss", Value: loss, Limit: 0, Blocking: true,
				Consumer: "repository-materializer", Operation: "restore-logical-tree", Proof: proof},
			{ID: "projection.unbound-entry", Value: unbound, Limit: 0, Blocking: true,
				Consumer: "repository-projector", Operation: "bind-content-object", Proof: proof},
			{ID: directObservedIndicator, Value: topology.ObservedDirect, Limit: -1, Blocking: false,
				Consumer: "repository-topology-classifier", Operation: "observe-direct-object-buckets", Proof: proof},
			{ID: directIndicator, Value: topology.Direct, Limit: 0, Blocking: true,
				Consumer: "radix-sharder", Operation: "split-object-bucket", Proof: proof},
			{ID: bootstrapDirectIndicator, Value: topology.ExemptDirect, Limit: 1, Blocking: true,
				Consumer: "github-actions", Operation: "preserve-workflow-discovery", Proof: proof},
			{ID: directUnboundIndicator, Value: unclassifiedDirect, Limit: 0, Blocking: true,
				Consumer: "repository-topology-classifier", Operation: "classify-direct-object-buckets", Proof: proof},
			{ID: mixedIndicator, Value: topology.Mixed, Limit: 0, Blocking: true,
				Consumer: "radix-sharder", Operation: "separate-branch-leaf", Proof: proof},
			{ID: "source.line-cap-debt", Value: lineDebt, Limit: 0, Blocking: false,
				Consumer: "logical-source-splitter", Operation: "split-before-storage", Proof: proof},
		},
	}
}
func requireBlockingZero(report evidence) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			details := blockingSubjectDetails(report.Subjects, metric.ID)
			if details != "" {
				return fmt.Errorf("blocking indicator %s=%d subjects=%s", metric.ID, metric.Value, details)
			}
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}

func blockingSubjectDetails(subjects []subject, indicatorID string) string {
	details := make([]string, 0)
	for _, item := range subjects {
		if item.Indicator != indicatorID || item.Physical == "" {
			continue
		}
		details = append(details, fmt.Sprintf("%s(%d>%d)", item.Physical, item.Value, item.Limit))
	}
	sort.Strings(details)
	return strings.Join(details, ",")
}
