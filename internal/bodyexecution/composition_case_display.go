package bodyexecution

import (
	"encoding/json"
	"fmt"
	"sort"
)

// CompositionCaseDisplay is a diagnostic lens for lowered workspace cases.
// It is excluded from case JSON and never changes an accepted input value.
type CompositionCaseDisplay struct {
	ActivityKeys    map[string]string
	SubmittedInputs map[string]CompositionSubmittedInput
}

type CompositionSubmittedInput struct {
	Key      string
	Activity string
}

func (display *CompositionCaseDisplay) activityKey(lowered string) string {
	if display != nil && display.ActivityKeys[lowered] != "" {
		return display.ActivityKeys[lowered]
	}
	return lowered
}

func (display *CompositionCaseDisplay) inputKey(node CompositionActivity, port string) string {
	name := display.activityKey(node.Name)
	if len(node.Inputs) == 0 {
		return name
	}
	return name + "." + port
}

func (graph compositionGraph) displayInputNames(display *CompositionCaseDisplay) (map[string]bool, map[string][]string) {
	declared := make(map[string]bool)
	available := make(map[string][]string)
	for _, node := range graph.nodes[:graph.count] {
		for _, slot := range node.inputSlots() {
			declared[node.inputKey(slot.Port)] = true
			if slot.From < 0 {
				available[node.Name] = append(available[node.Name], display.inputKey(node, slot.Port))
			}
		}
	}
	for activity := range available {
		sort.Strings(available[activity])
	}
	return declared, available
}

func (display *CompositionCaseDisplay) checkSubmittedInputs(inputs map[string]json.RawMessage,
	declared map[string]bool, available map[string][]string) error {
	var unknown []string
	for key := range inputs {
		if !declared[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	key := unknown[0]
	origin := display.SubmittedInputs[key]
	shown := origin.Key
	if shown == "" {
		shown = key
	}
	choices := available[origin.Activity]
	if len(choices) == 1 {
		return fmt.Errorf("input %q is not a caller port; use %q", shown, choices[0])
	}
	if len(choices) > 1 {
		return fmt.Errorf("input %q is not a caller port; available ports: %q", shown, choices)
	}
	return fmt.Errorf("input %q is not a caller port", shown)
}
