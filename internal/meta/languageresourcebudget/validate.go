package languageresourcebudget

import (
	"fmt"
	"reflect"
)

func ValidateReport(input Input, report Report) error {
	want := Evaluate(input, report.Case)
	if !reflect.DeepEqual(want, report) {
		return fmt.Errorf("RESOURCE_REPORT_NOT_DETERMINISTIC")
	}
	return nil
}
