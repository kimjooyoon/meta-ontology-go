package generator

func acceptanceFixture() SemanticIR {
	return SemanticIR{
		Package: "bootstrapgen",
		Entities: []Entity{
			{
				ID: "gooo://entity/source", Name: "Source", GoName: "Source",
			},
			{
				ID: "gooo://entity/artifact", Name: "Artifact", GoName: "Artifact",
			},
		},
		Activities: []Activity{
			{
				ID: "gooo://activity/compile", Name: "Compile", GoName: "Compile",
				Inputs:  []Port{{EntityID: "gooo://entity/source", Name: "source"}},
				Outputs: []Port{{EntityID: "gooo://entity/artifact", Name: "artifact"}},
				Slots:   []Slot{{ID: "gooo://slot/compile-implementation", Default: "return Artifact{}"}},
			},
			{
				ID: "gooo://activity/inspect", Name: "Inspect", GoName: "Inspect",
				Inputs:  []Port{{EntityID: "gooo://entity/artifact", Name: "artifact"}},
				Outputs: []Port{{EntityID: "gooo://entity/artifact", Name: "result"}},
				Slots:   []Slot{{ID: "gooo://slot/inspect-implementation", Default: "return artifact"}},
			},
		},
	}
}
