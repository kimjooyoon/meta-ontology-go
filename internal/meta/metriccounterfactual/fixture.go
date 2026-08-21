package metriccounterfactual

func ProjectRootPolicy() RootPolicy {
	return RootPolicy{
		CountsApplicability:   "OBSERVED",
		TopologyApplicability: "NOT_APPLICABLE",
		TopologyReason:        "ROOT_TOPOLOGY_EXEMPT",
		ReadmeRequirement:     "NOT_APPLICABLE",
	}
}

func BaselineManifest() (Manifest, error) {
	return SealManifest(Manifest{
		Schema: ManifestSchema,
		Files: []FileSpec{
			{
				Path:     "logic/rules.gooo",
				Language: "gooo",
				Content:  "entity source.\nrule derive(source).\n",
			},
			{
				Path:     "runtime/main.go",
				Language: "go",
				Content:  "package runtime\n\nfunc MainValue() int { return 1 }\n",
			},
			{
				Path:     "runtime/nested/existing.go",
				Language: "go",
				Content:  "package nested\n\nvar Existing = 1\n",
			},
		},
	})
}

func CounterfactualPlan() (Plan, error) {
	return SealPlan(Plan{
		Schema: PlanSchema,
		Mutations: []Mutation{
			{
				Kind:    "APPEND",
				Path:    "logic/rules.gooo",
				Content: "derive improved.\n",
			},
			{
				Kind:    "CREATE",
				Path:    "generated/deeper/new.go",
				Content: "package generated\n\nfunc Value() int { return 2 }\n",
			},
		},
	})
}
