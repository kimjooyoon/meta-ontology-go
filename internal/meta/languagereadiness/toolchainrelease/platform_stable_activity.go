package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const stableActivityID = "urn:gooo:example:total"

func stableActivityExamples() []languageSmokeCase {
	return []languageSmokeCase{
		{"stable-activity-english", "examples/stable-activity-identity/english/source.gooo.fixture", "examples/stable-activity-identity/english/cases.json", "Total", 4},
		{"stable-activity-korean", "examples/stable-activity-identity/korean/source.gooo.fixture", "examples/stable-activity-identity/korean/cases.json", "합계", 4},
	}
}

func smokeStableActivity(binary, work string, input BuildInput) error {
	common := ""
	for _, example := range stableActivityExamples() {
		directory := filepath.Join(work, example.name)
		selected := ""
		for _, replay := range []bool{false, true} {
			raw, err := runLanguageSmokeCommand(binary, directory, input, example, replay, commandOutput)
			if err != nil {
				return fmt.Errorf("TOOLCHAIN_RELEASE_STABLE_ACTIVITY %s: %w", example.name, err)
			}
			selected, err = validateLanguageSmoke(raw, example.total, replay, selected)
			if err != nil {
				return err
			}
			if err := validateStableActivitySmoke(raw, example.entry); err != nil {
				return err
			}
			if common != "" && common != selected {
				return fmt.Errorf("renaming the activity changed its generated program")
			}
			common = selected
		}
	}
	return nil
}

func validateStableActivitySmoke(raw []byte, entry string) error {
	var result struct {
		Composition bodyexecution.Composition        `json:"composition"`
		Runtime     bodyexecution.CompositionRuntime `json:"runtime"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return err
	}
	plan := result.Composition.Plan
	if entry == "" || plan.EntryActivity != entry || len(plan.Activities) != 1 {
		return fmt.Errorf("stable activity entry or declaration count differs")
	}
	activity := plan.Activities[0]
	if activity.Name != entry || activity.ID != stableActivityID || len(activity.Inputs) != 2 {
		return fmt.Errorf("stable activity name, identity or input count differs")
	}
	for i, port := range activity.Inputs {
		if port.Port != fmt.Sprintf("input%d", i) || port.EntityID != "urn:gooo:type:integer" || port.From != -1 {
			return fmt.Errorf("stable activity typed input differs")
		}
	}
	for _, boundary := range []string{"start", "end"} {
		marker := fmt.Sprintf("//gooo:generated:%s id=%q kind=\"activity\"", boundary, stableActivityID)
		if strings.Count(result.Composition.Source, marker) != 1 {
			return fmt.Errorf("stable activity generated marker differs")
		}
	}
	return validateStableActivityTraces(result.Runtime.Traces)
}

func validateStableActivityTraces(traces []bodyexecution.CompositionTrace) error {
	rows := [4][3]string{{"2", "3", "5"}, {"-7", "2", "0"}, {"0", "0", "0"}, {"9007199254740993", "7", "9007199254741000"}}
	if len(traces) != len(rows) {
		return fmt.Errorf("stable activity native trace count differs")
	}
	for i, trace := range traces {
		if trace.CaseIndex != i || len(trace.Deliveries) != 1 {
			return fmt.Errorf("stable activity native trace order differs")
		}
		delivery := trace.Deliveries[0]
		if delivery.ActivityID != stableActivityID || len(delivery.Inputs) != 2 || delivery.Fault != nil ||
			delivery.Passed == nil || !*delivery.Passed || string(delivery.Actual) != rows[i][2] || string(delivery.Expected) != rows[i][2] {
			return fmt.Errorf("stable activity native identity or exact result differs")
		}
		for j, input := range delivery.Inputs {
			if input.Port != fmt.Sprintf("input%d", j) || input.EntityID != "urn:gooo:type:integer" || string(input.Value) != rows[i][j] {
				return fmt.Errorf("stable activity native input differs")
			}
		}
	}
	return nil
}
