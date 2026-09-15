package rollbackfixedpoint

import (
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	if report.Schema != Schema || len(report.Coordinates) != totalCoordinates ||
		len(report.Indicators) != 8 || len(report.Proofs) != 3 ||
		!validDigest(report.ReportDigest) {
		return fmt.Errorf("rollback fixed-point report is incomplete")
	}
	if !reflect.DeepEqual(report, Build(report.Source)) {
		return fmt.Errorf("rollback fixed-point report does not replay")
	}
	return nil
}
