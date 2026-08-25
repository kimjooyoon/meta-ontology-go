package main

func countWitnesses(witnesses []subjectWitness) ledgerCounts {
	values := make(map[string]int)
	for _, witness := range witnesses {
		values["subjects"]++
		values["space:"+witness.Space]++
		values["language:"+witness.Language]++
		values["binding:"+witness.Meta.Kind]++
		values[witness.Space+"\x00"+witness.Meta.Kind]++
		values["source:applicable"] += witness.Meta.ApplicableIndicators
		values["source:not-applicable"] += witness.Meta.NotApplicableIndicators
		if witness.Space == "STORAGE_DIRECTORY" && witness.Path == "." {
			values["root-summary"] += witness.Meta.IndicatorCount - rootBaseIndicatorCount
		}
		if witness.Space == "STORAGE_DIRECTORY" && witness.Path == workflowDiscoveryPath {
			values["workflow-discovery"] += witness.Meta.NotApplicableIndicators
		}
	}
	return ledgerCounts{
		FileWitnesses: values["space:LOGICAL_FILE"], FunctionWitnesses: values["space:LOGICAL_FUNCTION"], GoFiles: values["language:go"], GoooFiles: values["language:gooo"],
		OtherFiles: values["language:other"], LogicalDirectories: values["space:LOGICAL_DIRECTORY"], StorageDirectories: values["space:STORAGE_DIRECTORY"],
		FileSourceBindings: values["LOGICAL_FILE\x00SOURCE_INDICATORS"], FunctionSourceBindings: values["LOGICAL_FUNCTION\x00SOURCE_INDICATORS"], StorageSourceBindings: values["STORAGE_DIRECTORY\x00SOURCE_INDICATORS"], RootSummaryIndicators: values["root-summary"],
		DerivedBindings: values["binding:DERIVED_OBSERVATION"] + values["binding:DERIVED_PROJECTION"], SubjectWitnesses: values["subjects"],
		SourceIndicatorsApplicable: values["source:applicable"], SourceIndicatorsNotApplicable: values["source:not-applicable"], WorkflowDiscoveryExemptions: values["workflow-discovery"],
	}
}
