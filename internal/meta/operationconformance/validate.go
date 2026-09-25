package operationconformance

import (
	"errors"
	"reflect"
)

func Validate(report Report, contractRaw []byte) error {
	if report.Schema != ReportSchema || report.ContractID != ContractID || report.OperationID != OperationID {
		return errors.New("SplitGo conformance report identity mismatch")
	}
	expected := Evaluate(contractRaw, report.Evidence)
	if !reflect.DeepEqual(report, expected) {
		return errors.New("SplitGo conformance report is not deterministic")
	}
	return nil
}
