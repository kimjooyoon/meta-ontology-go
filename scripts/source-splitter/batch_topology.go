package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func validateBatchTopology(report metricevidence.Report, plans []plannedSplit) error {
	added := map[string]int{}
	for _, item := range plans {
		added[item.plan.Directory] += len(item.plan.Parts) - 1
	}
	for _, directory := range report.Directories {
		count, selected := added[directory.Path]
		if !selected {
			continue
		}
		if directory.DirectFolders != 0 || directory.DirectFiles+count > report.Meta.Policy.MaxDirectEntries {
			return fmt.Errorf("batch split blocked: %s projects %d entries",
				directory.Path, directory.DirectFiles+count)
		}
		delete(added, directory.Path)
	}
	if len(added) != 0 {
		return fmt.Errorf("batch split topology evidence omits %d directories", len(added))
	}
	return nil
}
