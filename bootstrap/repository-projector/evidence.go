package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func topologyFailures(root string) (int, int, error) {
	direct, mixed := 0, 0
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || !entry.IsDir() {
			return walkErr
		}
		children, err := os.ReadDir(name)
		if err != nil {
			return err
		}
		if len(children) > 10 {
			direct++
		}
		hasDirectory, hasFile := false, false
		for _, child := range children {
			hasDirectory = hasDirectory || child.IsDir()
			hasFile = hasFile || !child.IsDir()
		}
		if hasDirectory && hasFile {
			mixed++
		}
		return nil
	})
	return direct, mixed, err
}
func buildEvidence(sha string, model manifest, objects int,
	loss, direct, mixed int) evidence {
	unbound, lineDebt := 0, 0
	subjects := make([]subject, 0)
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
	proof := "axiomatic-foundation"
	return evidence{
		Schema: "gooo.repository-projection-evidence.v1", SourceSHA: sha,
		TrackedFiles: len(model.Entries), Objects: objects, Subjects: subjects,
		Indicators: []indicator{
			{ID: "projection.roundtrip-loss", Value: loss, Limit: 0, Blocking: true,
				Consumer: "repository-materializer", Operation: "restore-logical-tree", Proof: proof},
			{ID: "projection.unbound-entry", Value: unbound, Limit: 0, Blocking: true,
				Consumer: "repository-projector", Operation: "bind-content-object", Proof: proof},
			{ID: "storage.direct-entry", Value: direct, Limit: 0, Blocking: true,
				Consumer: "radix-sharder", Operation: "split-object-bucket", Proof: proof},
			{ID: "storage.mixed-kind", Value: mixed, Limit: 0, Blocking: true,
				Consumer: "radix-sharder", Operation: "separate-branch-leaf", Proof: proof},
			{ID: "source.line-cap-debt", Value: lineDebt, Limit: 0, Blocking: false,
				Consumer: "logical-source-splitter", Operation: "split-before-storage", Proof: proof},
		},
	}
}
func requireBlockingZero(report evidence) error {
	for _, metric := range report.Indicators {
		if metric.Blocking && metric.Value > metric.Limit {
			return fmt.Errorf("blocking indicator %s=%d", metric.ID, metric.Value)
		}
	}
	return nil
}
