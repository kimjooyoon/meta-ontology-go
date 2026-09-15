package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func fillSelection(result PlanResult, selected []selectedPath, frontier workfrontier.Result) PlanResult {
	workByCommand := map[string]string{}
	for i, pathID := range frontier.SelectedIDs {
		if i < len(frontier.WorkIDs) {
			workByCommand[pathID] = frontier.WorkIDs[i]
		}
	}
	for _, entry := range selected {
		if entry.guard {
			result.SelectedGuardCommandIDs = append(result.SelectedGuardCommandIDs, entry.command.ID)
		} else {
			result.SelectedCommandIDs = append(result.SelectedCommandIDs, entry.command.ID)
		}
		result.SelectedWorkIDs = append(result.SelectedWorkIDs, workByCommand[entry.command.ID])
	}
	return result
}
