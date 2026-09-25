package promotioncontinuity

import (
	"errors"
	"reflect"
)

func Validate(report Report) error {
	if report.Schema != Schema {
		return errors.New("promotion continuity schema is not canonical")
	}
	expected := Evaluate(report.Source.ExpectedHeadSHA, report.Source.Guard, report.Source.Recovery)
	if !reflect.DeepEqual(report, expected) {
		return errors.New("promotion continuity report is not the deterministic evaluation")
	}
	return nil
}
