package generator

import "fmt"

func validateDeclaredRegions(ir SemanticIR, markers parsedMarkers) error {
	declared := make(map[string]string, len(ir.Entities)+len(ir.Activities))
	for _, entity := range ir.Entities {
		declared[entity.ID] = "entity"
	}
	for _, activity := range ir.Activities {
		declared[activity.ID] = "activity"
	}
	for _, region := range markers.Regions {
		expected, exists := declared[region.ID]
		if !exists {
			continue
		}
		if region.Kind != expected {
			return fmt.Errorf("generator: generated region %q changes kind from %q to %q", region.ID, region.Kind, expected)
		}
	}
	return nil
}
