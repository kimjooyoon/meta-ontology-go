package reproducibilitysemantics

import (
	"fmt"
	"reflect"
)

func ValidateJudgment(sourcePath, headSHA string, source []byte, receipt Receipt, judgment Judgment) error {
	want := Judge(sourcePath, headSHA, source, receipt)
	if judgment.Decision != StatusDischarged || judgment.Reason != "NON_IDENTITY_EXHIBITED" {
		return fmt.Errorf("reproducibility semantics judgment failed: %s", judgment.Reason)
	}
	if !reflect.DeepEqual(judgment, want) {
		return fmt.Errorf("reproducibility semantics judgment is not deterministic")
	}
	return nil
}
