package roundtrip

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// CheckLocality reports generated regions that changed outside the accepted
// semantic locality. Text outside generated markers is deliberately ignored:
// it belongs to handwritten implementation slots or other source owners.
func CheckLocality(input LocalityInput) Report {
	before, beforeErr := parseGeneratedFile(input.Before)
	after, afterErr := parseGeneratedFile(input.After)
	var report Report
	if beforeErr != nil {
		report.add(Violation{Rule: RuleMarker, Path: "before-go", Detail: beforeErr.Error()})
	}
	if afterErr != nil {
		report.add(Violation{Rule: RuleMarker, Path: "after-go", Detail: afterErr.Error()})
	}
	if !report.OK() {
		report.normalize()
		return report
	}
	allowed := make(map[semantic.ID]struct{}, len(input.AllowedIDs))
	for _, rawID := range input.AllowedIDs {
		id, err := semantic.ParseIdentity(rawID.String())
		if err != nil {
			report.add(Violation{Rule: RuleSnapshot, Path: "allowed-id", Identity: rawID.String(), Detail: err.Error()})
			continue
		}
		allowed[id] = struct{}{}
	}
	if !report.OK() {
		report.normalize()
		return report
	}
	ids := regionIDs(before.Regions, after.Regions)
	for _, id := range ids {
		old, oldExists := before.Regions[id]
		current, currentExists := after.Regions[id]
		if oldExists && currentExists && sameRegion(old, current) {
			continue
		}
		if _, permitted := allowed[id]; permitted {
			continue
		}
		detail := "generated region changed outside semantic locality"
		if !oldExists {
			detail = "generated region was added outside semantic locality"
		}
		if !currentExists {
			detail = "generated region was removed outside semantic locality"
		}
		report.add(Violation{Rule: RuleLocality, Path: "generated-go", Identity: id.String(), Detail: detail})
	}
	report.normalize()
	return report
}
func sameRegion(left, right generatedRegion) bool {
	return left.Kind == right.Kind && string(left.Body) == string(right.Body)
}
