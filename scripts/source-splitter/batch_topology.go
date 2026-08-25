package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func batchTopologyDeferrals(report metricevidence.Report, plans []plannedSplit) (int, error) {
	added, subjects := map[string]int{}, map[string]int{}
	for _, item := range plans {
		added[item.plan.Directory] += len(item.plan.Parts) - 1
		subjects[item.plan.Directory]++
	}
	deferred := 0
	for _, directory := range report.Directories {
		count, selected := added[directory.Path]
		if !selected {
			continue
		}
		if directory.DirectFolders != 0 || directory.DirectFiles+count > report.Meta.Policy.MaxDirectEntries {
			deferred += subjects[directory.Path]
		}
		delete(added, directory.Path)
	}
	if len(added) != 0 {
		return 0, fmt.Errorf("batch split topology evidence omits %d directories", len(added))
	}
	return deferred, nil
}
