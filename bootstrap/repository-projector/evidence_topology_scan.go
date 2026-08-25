package main

import (
	"os"
	"path/filepath"
	"sort"
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
