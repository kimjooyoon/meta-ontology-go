package gosplit

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func ValidateTopology(report metricevidence.Report, plan Plan) error {
	for _, directory := range report.Directories {
		if directory.Path != plan.Directory {
			continue
		}
		added := len(plan.Parts) - 1
		if directory.DirectFolders != 0 {
			return planBlocker(DirectoryMixed, plan, fmt.Errorf("%w: %s contains folders", ErrBlocked, plan.Directory))
		}
		if directory.DirectFiles+added > report.Meta.Policy.MaxDirectEntries {
			return planBlocker(DirectoryFull, plan,
				fmt.Errorf("%w: %s projects %d entries", ErrBlocked, plan.Directory, directory.DirectFiles+added))
		}
		return nil
	}
	return planBlocker(DirectoryUnobserved, plan, fmt.Errorf("metric evidence omits directory %q", plan.Directory))
}
