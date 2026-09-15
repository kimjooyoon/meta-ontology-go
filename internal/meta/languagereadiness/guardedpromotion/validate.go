package guardedpromotion

import (
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	if report.Schema != Schema {
		return fmt.Errorf("guarded promotion schema %q is unknown", report.Schema)
	}
	if !validDigest(report.ReportDigest) {
		return fmt.Errorf("guarded promotion report digest is invalid")
	}
	if len(report.Coordinates) != 12 {
		return fmt.Errorf("guarded promotion coordinates = %d, want 12", len(report.Coordinates))
	}
	if len(report.Indicators) != 8 {
		return fmt.Errorf("guarded promotion indicators = %d, want 8", len(report.Indicators))
	}
	if len(report.Proofs) != 3 {
		return fmt.Errorf("guarded promotion proofs = %d, want 3", len(report.Proofs))
	}
	expected := Build(report.Source)
	if !reflect.DeepEqual(report, expected) {
		return fmt.Errorf("guarded promotion report does not replay")
	}
	return nil
}
