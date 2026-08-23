package languagegointeroperation

import "github.com/kimjooyoon/meta-ontology-go/internal/generator"

func generatorFixture(id string) (generator.SemanticIR, bool) {
	source, artifact := entity("entity:source", "Source"), entity("entity:artifact", "Artifact")
	model := generator.SemanticIR{Package: "interopgen"}
	switch id {
	case "single-entity":
		model.Entities = []generator.Entity{source}
	case "two-entities":
		model.Entities = []generator.Entity{source, artifact}
	case "builtin-output":
		model.Activities = []generator.Activity{builtinOutputActivity()}
	case "entity-flow":
		model.Entities = []generator.Entity{source, artifact}
		model.Activities = []generator.Activity{flowActivity("activity:compile", "Compile", source, artifact)}
	case "two-inputs":
		model = twoInputModel(source, artifact)
	case "two-activities":
		model = twoActivityModel(source, artifact)
	case "explicit-slot":
		model = explicitSlotModel(source, artifact)
	case "ordered-pipeline":
		model = orderedPipelineModel(source, artifact)
	default:
		return generator.SemanticIR{}, false
	}
	return model, true
}

func entity(id, name string) generator.Entity {
	return generator.Entity{ID: id, Name: name, GoName: name}
}

func port(id, name string, value generator.Entity) generator.Port {
	return generator.Port{ID: id, Name: name, GoName: name, EntityID: value.ID, GoType: value.GoName}
}

func flowActivity(id, name string, input, output generator.Entity) generator.Activity {
	return generator.Activity{ID: id, Name: name, GoName: name,
		Inputs:  []generator.Port{port("port:"+name+":in", "input", input)},
		Outputs: []generator.Port{port("port:"+name+":out", "output", output)}}
}

func builtinOutputActivity() generator.Activity {
	return generator.Activity{ID: "activity:render", Name: "Render", GoName: "Render",
		Outputs: []generator.Port{{ID: "port:render:result", Name: "result", GoName: "result", GoType: "string"}}}
}

func twoInputModel(source, artifact generator.Entity) generator.SemanticIR {
	config := entity("entity:config", "Config")
	activity := flowActivity("activity:compile", "Compile", source, artifact)
	activity.Inputs = append(activity.Inputs, port("port:compile:config", "config", config))
	return generator.SemanticIR{Package: "interopgen", Entities: []generator.Entity{source, config, artifact}, Activities: []generator.Activity{activity}}
}

func twoActivityModel(source, artifact generator.Entity) generator.SemanticIR {
	compile := flowActivity("activity:compile", "Compile", source, artifact)
	publish := generator.Activity{ID: "activity:publish", Name: "Publish", GoName: "Publish", Inputs: []generator.Port{port("port:publish:artifact", "artifact", artifact)}}
	return generator.SemanticIR{Package: "interopgen", Entities: []generator.Entity{source, artifact}, Activities: []generator.Activity{compile, publish}}
}

func explicitSlotModel(source, artifact generator.Entity) generator.SemanticIR {
	activity := flowActivity("activity:compile", "Compile", source, artifact)
	activity.Slots = []generator.Slot{{ID: "slot:compile", Name: "implementation", Default: "return Artifact{}"}}
	return generator.SemanticIR{Package: "interopgen", Entities: []generator.Entity{source, artifact}, Activities: []generator.Activity{activity}}
}

func orderedPipelineModel(source, artifact generator.Entity) generator.SemanticIR {
	ir := entity("entity:ir", "IR")
	lower := flowActivity("activity:lower", "Lower", source, ir)
	emit := flowActivity("activity:emit", "Emit", ir, artifact)
	return generator.SemanticIR{Package: "interopgen", Entities: []generator.Entity{source, ir, artifact}, Activities: []generator.Activity{lower, emit}}
}
