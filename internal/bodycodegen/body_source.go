package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type preparedBody struct {
	activity                                  *syntax.ActivityDecl
	packageName, activityID, outputType, body string
	parameters                                []InputParameter
	records                                   []RecordType
	base                                      generatedRoute
}

func prepareActivityBody(filename string, source []byte, activityName string) (preparedBody, error) {
	var p preparedBody
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() {
		return p, fmt.Errorf("parse .gooo source: %w", diagnostics.Error())
	}
	if file == nil || file.Package == nil {
		return p, fmt.Errorf(".gooo source has no package declaration")
	}
	var err error
	p.activity, err = sourceBodyActivity(file, activityName)
	if err != nil {
		return p, err
	}
	model, records, err := resolveBodyModel(file)
	if err != nil {
		return p, err
	}
	p.parameters, err = sourceBodyParameters(p.activity, records)
	if err != nil {
		return p, err
	}
	var ok bool
	p.outputType, ok = bodyEntityType(p.activity.Output, records)
	if !ok {
		return p, fmt.Errorf("activity %q output entity %q is outside the pure value profile", activityName, p.activity.Output)
	}
	p.records = activityRecords(records, p.parameters, p.outputType, p.activity.ValueProgram)
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == activityName {
			p.activityID = string(node.ID)
			break
		}
	}
	if p.activityID == "" {
		return p, fmt.Errorf("activity %q has no stable semantic identity", activityName)
	}
	p.packageName = file.Package.Name
	p.body, err = rewriteLetDeclarations(p.activity.ValueProgram)
	if err != nil {
		return p, err
	}
	p.base, err = p.generate(preserveRoute)
	return p, err
}

func sourceBodyActivity(file *syntax.File, name string) (*syntax.ActivityDecl, error) {
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != name {
			continue
		}
		if !activity.ValueProgramPresent || activity.ValueProgram == "" {
			return nil, fmt.Errorf("activity %q has no computes body", name)
		}
		return activity, nil
	}
	return nil, fmt.Errorf("activity %q was not found", name)
}

func sourceBodyParameters(activity *syntax.ActivityDecl, records []RecordType) ([]InputParameter, error) {
	if len(activity.Inputs) < 1 || len(activity.Inputs) > 16 {
		return nil, fmt.Errorf("activity %q requires 1..16 inputs, got %d", activity.Name, len(activity.Inputs))
	}
	parameters := make([]InputParameter, len(activity.Inputs))
	for i, input := range activity.Inputs {
		inputType, ok := bodyEntityType(input.Name, records)
		if !ok {
			return nil, fmt.Errorf("activity %q input entity %q is outside the pure value profile", activity.Name, input.Name)
		}
		name := "input"
		if len(activity.Inputs) > 1 {
			name = fmt.Sprintf("input%d", i)
		}
		parameters[i] = InputParameter{Name: name, Type: inputType}
	}
	return parameters, nil
}

func (p preparedBody) generate(route string) (generatedRoute, error) {
	return generateRouteParameters(p.packageName, p.activity.Name, p.activityID,
		p.parameters, p.outputType, p.body, route, p.records...)
}
