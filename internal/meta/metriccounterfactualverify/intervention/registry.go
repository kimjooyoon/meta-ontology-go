package metricintervention

import "fmt"

type Dimension struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Unit          string `json:"unit"`
	Family        string `json:"family"`
	Trilemma      string `json:"trilemma"`
	MetaOperation string `json:"meta_operation"`
}

type Registry struct {
	Schema     string      `json:"schema"`
	Dimensions []Dimension `json:"dimensions"`
}

func DefaultRegistry() Registry {
	return Registry{Schema: RegistrySchema, Dimensions: []Dimension{
		stateDimension("direct_folders", "folders"),
		stateDimension("direct_files", "files"),
		stateDimension("recursive_folders", "folders"),
		stateDimension("recursive_files", "files"),
		stateDimension("go_files", "files"),
		stateDimension("gooo_files", "files"),
		stateDimension("go_lines", "lines"),
		stateDimension("gooo_lines", "lines"),
		eventDimension("changed_files", "files"),
		eventDimension("changed_directories", "directories"),
	}}
}

func stateDimension(id, unit string) Dimension {
	return Dimension{ID: id, Kind: "STATE", Unit: unit, Family: "COHERENCE", Trilemma: "COHERENCE", MetaOperation: "project-algebraic-root-state"}
}

func eventDimension(id, unit string) Dimension {
	return Dimension{ID: id, Kind: "EVENT", Unit: unit, Family: "REGRESSION", Trilemma: "REGRESS", MetaOperation: "observe-counterfactual-boundary"}
}

func ValidateRegistry(registry Registry) error {
	if registry.Schema != RegistrySchema || len(registry.Dimensions) != 10 {
		return fmt.Errorf("metric dimension registry contract is invalid")
	}
	seen := make(map[string]bool, len(registry.Dimensions))
	for _, dimension := range registry.Dimensions {
		if dimension.ID == "" || seen[dimension.ID] || (dimension.Kind != "STATE" && dimension.Kind != "EVENT") {
			return fmt.Errorf("metric dimension %q is invalid", dimension.ID)
		}
		seen[dimension.ID] = true
	}
	return nil
}
