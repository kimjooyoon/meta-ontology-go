package proposalcompat

import (
	"fmt"
	"reflect"
)

func Validate(bundle Bundle, expectedHead string) error {
	legacy := bundle.Legacy
	expectedLegacy := legacy
	expectedLegacy.ReportDigest = ""
	expectedLegacy = sealLegacy(expectedLegacy)
	switch {
	case legacy.Schema != LegacySchema || legacy.CurrentHeadSHA != expectedHead:
		return fmt.Errorf("compatibility legacy identity is not exact")
	case legacy.Decision != DecisionPass || legacy.Summary.Satisfied != 8 ||
		legacy.Summary.Total != 8 || legacy.Summary.Unresolved != 0 ||
		legacy.Summary.RepositoryWrites != 0:
		return fmt.Errorf("compatibility legacy contract is not ready")
	case !reflect.DeepEqual(legacy, expectedLegacy):
		return fmt.Errorf("compatibility legacy digest diverged")
	}
	report := bundle.Receipt
	if report.Schema != Schema || report.Source.ExpectedHeadSHA != expectedHead ||
		len(report.Coordinates) != totalCoordinates || len(report.Indicators) != 8 ||
		len(report.Proofs) != 3 || report.Decision != DecisionPass {
		return fmt.Errorf("compatibility receipt is incomplete")
	}
	if report.Source.TargetReportDigest != legacy.ReportDigest ||
		report.Source.TargetFileSHA256 != digestBytes(EncodeLegacy(legacy)) {
		return fmt.Errorf("compatibility target linkage diverged")
	}
	if !reflect.DeepEqual(report, buildReceipt(report.Source)) {
		return fmt.Errorf("compatibility receipt does not replay")
	}
	return nil
}
