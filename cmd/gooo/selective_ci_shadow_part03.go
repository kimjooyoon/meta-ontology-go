package main

func readSelectiveCIShadowFiles(options selectiveCIShadowOptions, reader SourceReader) (shadowInputFiles, string) {
	read := func(name, filename string) ([]byte, string) {
		data, err := reader.ReadFile(filename)
		if err != nil {
			return nil, name
		}
		return data, ""
	}
	base, missing := read("base_snapshot", options.baseSnapshot)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	head, missing := read("head_snapshot", options.headSnapshot)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	plan, missing := read("plan_input", options.planInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	evidence, missing := read("evidence_input", options.evidenceInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	lane, missing := read("lane_input", options.laneInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	return shadowInputFiles{baseSnapshot: base, headSnapshot: head, planInput: plan, evidenceInput: evidence, laneInput: lane}, ""
}
func newSelectiveCIShadowOutput() selectiveCIShadowOutput {
	return selectiveCIShadowOutput{
		SchemaVersion:       selectiveCIShadowSchemaVersion,
		Command:             "selective-ci shadow",
		ExecutionAuthorized: false,
		ShadowOnly:          true,
		ChangedSemanticIDs:  []string{},
		SelectedCommands:    []shadowCommandSpec{},
		SelectedGuards:      []shadowCommandSpec{},
		SelectedWorkIDs:     []string{},
		ResourceReceipts:    []shadowResourceReceipt{},
	}
}
