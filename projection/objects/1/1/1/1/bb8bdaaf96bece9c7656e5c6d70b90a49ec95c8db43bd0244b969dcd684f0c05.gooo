package pressureshadow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

func (input *Input) UnmarshalJSON(data []byte) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	type wire Input
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	if err := validateSyntax(Input(decoded)); err != nil {
		return err
	}
	*input = Input(decoded)
	return nil
}
func DecodeInput(data []byte) (Input, error) {
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return Input{}, err
	}
	return input, nil
}
func validateSyntax(input Input) error {
	if input.Schema != SchemaVersion || input.Selector.SchemaVersion != workfrontier.SchemaVersion {
		return fmt.Errorf("invalid schema")
	}
	seenPressures := make(map[string]struct{}, len(input.Selector.Pressures))
	for _, pressure := range input.Selector.Pressures {
		if err := checkUniqueID(seenPressures, pressureID(pressure), "pressure"); err != nil {
			return err
		}
	}
	seenStates := make(map[string]struct{}, len(input.Selector.States))
	for _, state := range input.Selector.States {
		id := stableID(state.ObligationID, state.ID)
		if err := checkUniqueID(seenStates, id, "obligation state"); err != nil {
			return err
		}
	}
	seenPaths := make(map[string]struct{}, len(input.Selector.Paths))
	for _, path := range input.Selector.Paths {
		id := pathID(path)
		if !validIDs(path.RequiredPressureIDs) {
			return fmt.Errorf("invalid path ID")
		}
		if err := checkUniqueID(seenPaths, id, "path"); err != nil {
			return err
		}
	}
	seenRows := make(map[string]struct{}, len(input.PathCoverage))
	for _, row := range input.PathCoverage {
		if err := checkUniqueID(seenRows, row.PathID, "path coverage"); err != nil {
			return err
		}
		if _, err := pressurecoverage.CanonicalInputBytes(row.Coverage); err != nil {
			return fmt.Errorf("invalid pressure coverage: %w", err)
		}
	}
	return nil
}
