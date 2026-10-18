package workspaceexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// InputTemplate projects unbound ports in the prepared Gooo graph to their
// authored package keys. Zero values are editable shapes, not inferred cases.
func InputTemplate(ctx context.Context, program Program) ([]byte, error) {
	plan, err := bodyexecution.InspectComposition(
		ctx, "workspace-lowered.gooo", []byte(program.Source), program.Entry.LoweredName,
	)
	if err != nil {
		return nil, err
	}
	byName := make(map[string]ActivityRef, len(program.Activities))
	for _, activity := range program.Activities {
		if activity.LoweredName == "" || activity.PackagePath == "" || activity.Activity == "" {
			return nil, fmt.Errorf("workspace input template has an incomplete activity origin")
		}
		if _, exists := byName[activity.LoweredName]; exists {
			return nil, fmt.Errorf("workspace input template has ambiguous activity origins")
		}
		byName[activity.LoweredName] = activity
	}
	byID := make(map[string]bodyexecution.CompositionActivity, len(plan.Plan.Activities))
	for _, activity := range plan.Plan.Activities {
		if activity.ID == "" || activity.Name == "" {
			return nil, fmt.Errorf("workspace input template has an incomplete lowered activity")
		}
		if _, exists := byID[activity.ID]; exists {
			return nil, fmt.Errorf("workspace input template has ambiguous lowered activity IDs")
		}
		byID[activity.ID] = activity
	}
	for index := range plan.CallerInputs {
		input := &plan.CallerInputs[index]
		lowered, found := byID[input.ActivityID]
		if !found {
			return nil, fmt.Errorf("workspace input template cannot resolve caller activity %q", input.ActivityID)
		}
		original, found := byName[lowered.Name]
		if !found {
			return nil, fmt.Errorf("workspace input template cannot resolve original activity for %q", lowered.Name)
		}
		key := packageActivityKey(original.PackagePath, original.Activity)
		switch input.Key {
		case lowered.Name:
		case lowered.Name + "." + input.Port:
			key += "." + input.Port
		default:
			return nil, fmt.Errorf("workspace input template has an unexpected caller key %q", input.Key)
		}
		input.Key = key
	}
	return bodyexecution.CompositionInputTemplate(plan)
}
