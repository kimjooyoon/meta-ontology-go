package operationconformance

func resolveOracle(indicator string, observation map[string]any) Decision {
	complete, ok := boolValue(observation, "evidence_complete")
	if !ok || !complete {
		return DecisionUnknown
	}
	switch indicator {
	case fixedIndicators[0].ID:
		mode, modeOK := stringValue(observation, "replacement_mode")
		writes, writesOK := intValue(observation, "writes_outside_declared_targets")
		return compare(modeOK && writesOK && mode == "ATOMIC_RENAME_SAME_FILESYSTEM" && writes == 0)
	case fixedIndicators[1].ID:
		return compare(equalFields(observation, "build_context_digest_before", "build_context_digest_after") &&
			equalFields(observation, "selected_file_set_digest_before", "selected_file_set_digest_after"))
	case fixedIndicators[2].ID:
		return compare(equalFields(observation, "header_digest_before", "header_digest_after"))
	case fixedIndicators[3].ID:
		return compare(equalFields(observation, "import_set_union_digest_before", "import_set_union_digest_after"))
	case fixedIndicators[4].ID:
		return compare(equalFields(observation, "initialization_units_digest_before", "initialization_units_digest_after"))
	case fixedIndicators[5].ID:
		selected, selectedOK := intValue(observation, "selected_file_count")
		parsed, parsedOK := intValue(observation, "parsed_file_count")
		count, countOK := intValue(observation, "package_name_count")
		names, namesOK := stringsValue(observation, "package_names")
		return compare(selectedOK && parsedOK && countOK && namesOK &&
			selected == parsed && count == 1 && len(names) == 1 && names[0] != "")
	default:
		return DecisionUnknown
	}
}

func compare(value bool) Decision {
	if value {
		return DecisionPass
	}
	return DecisionFail
}
