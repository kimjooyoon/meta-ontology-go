package packageruntime

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

func sourceDeclarations(packagePath, filename string, declarations []syntax.Declaration) (
	[]string, []Export, []EntryPlan,
) {
	names := make([]string, 0, len(declarations))
	exports := make([]Export, 0, len(declarations))
	activities := make([]EntryPlan, 0)
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			names = append(names, value.Name)
			exports = append(exports, Export{Name: value.Name, Kind: "entity", ID: value.ID})
		case *syntax.ActivityDecl:
			names = append(names, value.Name)
			activities = append(activities, activityPlan(packagePath, filename, value))
			inputTypes := make([]string, len(value.Inputs))
			for index, input := range value.Inputs {
				inputTypes[index] = input.Name
			}
			outputType := value.Output
			if outputType == "" {
				outputType = value.Result.Name
			}
			exports = append(exports, Export{Name: value.Name, Kind: "activity", ID: value.ID, InputTypes: inputTypes, OutputType: outputType})
		}
	}
	return names, exports, activities
}

func activityPlan(packagePath, filename string, activity *syntax.ActivityDecl) EntryPlan {
	parameters := activity.Inputs
	if parameters == nil {
		parameters = activity.Parameters
	}
	inputs := make([]string, len(parameters))
	for index, parameter := range parameters {
		inputs[index] = parameter.Name
	}
	output := activity.Output
	if output == "" {
		output = activity.Result.Name
	}
	return EntryPlan{
		PackagePath: packagePath, Source: filename, Activity: activity.Name,
		Inputs: inputs, Output: output,
	}
}
