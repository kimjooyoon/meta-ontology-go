package packageruntime

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

func sourceDeclarations(packagePath, filename string, declarations []syntax.Declaration) (
	[]string, []EntryPlan,
) {
	names := make([]string, 0, len(declarations))
	activities := make([]EntryPlan, 0)
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			names = append(names, value.Name)
		case *syntax.ActivityDecl:
			names = append(names, value.Name)
			activities = append(activities, activityPlan(packagePath, filename, value))
		}
	}
	return names, activities
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
