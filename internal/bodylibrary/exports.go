package bodylibrary

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// StageExport selects semantic identity; Name is only the Go API presentation.
type StageExport struct {
	ActivityID string `json:"activity_id"`
	Name       string `json:"name"`
}

func resolveStageExports(plan bodyexecution.CompositionPlan, exports []StageExport) (map[string]string, string, error) {
	functions := make(map[string]string, len(plan.Activities))
	for _, activity := range plan.Activities {
		functions[activity.ID] = activity.GoFunction
	}
	names, used := map[string]string{}, map[string]bool{"Evaluate": true}
	var origins strings.Builder
	for _, export := range exports {
		if !token.IsIdentifier(export.Name) || !ast.IsExported(export.Name) || used[export.Name] {
			return nil, "", fmt.Errorf("stage export name %q must be a distinct exported Go identifier", export.Name)
		}
		function, ok := functions[export.ActivityID]
		if !ok || names[function] != "" {
			return nil, "", fmt.Errorf("stage export %q must select one distinct planned activity ID", export.ActivityID)
		}
		names[function], used[export.Name] = export.Name, true
		fmt.Fprintf(&origins, "// Exported stage name=%q activity=%q scope=authored-body-inputs\n", export.Name, export.ActivityID)
	}
	return names, origins.String(), nil
}
