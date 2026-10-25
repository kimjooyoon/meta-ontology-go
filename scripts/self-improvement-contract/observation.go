package main

import "slices"

func readOnlyObservation(model contractModel) bool {
	const inputID = "gooo://meta/language-experiment-receipt"
	const outputID = "gooo://meta/read-only-improvement-input"
	const observeID = "meta://activity/observe"
	if model.EntityByID[inputID] == "" || model.EntityByID[outputID] == "" ||
		!activityMatches(model.Activities[observeID], []string{inputID}, outputID) {
		return false
	}
	for id, activity := range model.Activities {
		if id == observeID {
			continue
		}
		if activity.Output == outputID {
			return false
		}
		if slices.Contains(activity.Inputs, outputID) {
			return false
		}
	}
	return true
}
