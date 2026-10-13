package toolchainrelease

import "encoding/json"

// samePackageProgramSource reads an older program at its recorded origin depth.
// An absent optional annotation stays absent; recorded origin claims and every
// other program member still have to match the current source-derived program.
func samePackageProgramSource(recorded, current json.RawMessage) bool {
	var prior, prepared map[string]json.RawMessage
	if json.Unmarshal(recorded, &prior) != nil || prior == nil ||
		json.Unmarshal(current, &prepared) != nil || prepared == nil {
		return false
	}
	if _, present := prior["condition_origins"]; !present {
		delete(prepared, "condition_origins")
	}
	comparison, err := json.Marshal(prepared)
	return err == nil && samePackageSourceValue(recorded, comparison)
}
