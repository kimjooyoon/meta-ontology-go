package bodyexecution

import "fmt"

func (query NativeOriginQuery) hasSubjectSelection() bool {
	return query.RootActivityID != "" || query.ActivityID != "" || query.FieldID != ""
}

func validateOriginSubjectSelection(query NativeOriginQuery) error {
	if query.hasSubjectSelection() && !isFieldOriginKind(query.Kind) {
		return fmt.Errorf("origin subject IDs require field-writes, field-uses or field-input-locations")
	}
	return nil
}

// Keep event and assignment numbers from the producer. Only the page's row
// ordinal changes after selection; lineage columns still carry full cases.
func (page *nativeOriginPage) addField(row any, root, activity, field string) {
	page.sourceRows++
	query := page.query
	if query.RootActivityID != "" && query.RootActivityID != root ||
		query.ActivityID != "" && query.ActivityID != activity ||
		query.FieldID != "" && query.FieldID != field {
		return
	}
	page.add(row)
}
