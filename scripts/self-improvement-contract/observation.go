package main

import "slices"

func readOnlyObservation(model contractModel) bool {
	if model.Entities["LanguageExperimentReceipt"] != "gooo://meta/language-experiment-receipt" ||
		model.Entities["ReadOnlyImprovementInput"] != "gooo://meta/read-only-improvement-input" ||
		!activityMatches(model.Activities["Observe"], []string{"LanguageExperimentReceipt"}, "ReadOnlyImprovementInput") {
		return false
	}
	for name, activity := range model.Activities {
		if name == "Observe" {
			continue
		}
		if activity.Output == "ReadOnlyImprovementInput" {
			return false
		}
		if slices.Contains(activity.Inputs, "ReadOnlyImprovementInput") {
			return false
		}
	}
	return true
}
