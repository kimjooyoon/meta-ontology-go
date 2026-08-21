package adapter

import (
	"reflect"
)

const (
	OracleNW001   = "ORACLE-NW-001"
	OracleNW002   = "ORACLE-NW-002"
	OracleNW003   = "ORACLE-NW-003"
	OracleNW004   = "ORACLE-NW-004"
	OracleNW005   = "ORACLE-NW-005"
	OracleNW006   = "ORACLE-NW-006"
	OracleFAIL001 = "ORACLE-FAIL-001"
	OracleFAIL002 = "ORACLE-FAIL-002"
	OraclePASS001 = "ORACLE-PASS-001"
	OracleID001   = "ORACLE-ID-001"
)

// NoWriteEvidenceError is a stable observer-oracle rejection.
type NoWriteEvidenceError struct {
	Code   string
	Detail string
}

func (e NoWriteEvidenceError) Error() string { return e.Code + ": " + e.Detail }
func oracleError(code, detail string) NoWriteEvidenceError {
	return NoWriteEvidenceError{Code: code, Detail: detail}
}

// VerifyNoWrite accepts only a trusted observer capture bound to this request.
func (o *NoWriteObservation) VerifyNoWrite(request Request) error {
	if o == nil {
		return oracleError(OracleNW001, "observer evidence is required")
	}
	if o.stamp == nil {
		return oracleError(OracleNW003, "observer-owned capture marker is missing")
	}
	if o.Binding != requestObservationBinding(request) {
		return oracleError(OracleID001, "observer binding does not match request")
	}
	if o.stamp.digest != observationSeal(*o) {
		return oracleError(OracleNW003, "observer seal does not match captured evidence")
	}
	if err := validateVerifiedWorkflow(o.Workflow); err != nil {
		return err
	}
	if err := validateObserverMutation(*o); err != nil {
		return err
	}
	if err := validateObservation(*o); err != nil {
		return err
	}
	if err := comparePrimary(o.Before.Source, o.After.Source, "source"); err != nil {
		return err
	}
	if err := comparePrimary(o.Before.Output, o.After.Output, "output"); err != nil {
		return err
	}
	if o.Before.Temp.Root != o.After.Temp.Root || o.Before.Temp.Digest != o.After.Temp.Digest {
		return oracleError(OracleNW006, "temporary artifact snapshot changed")
	}
	current, err := captureState(o.Paths)
	if err != nil || !reflect.DeepEqual(current, o.After) {
		return oracleError(OracleNW002, "observer trace is stale")
	}
	return nil
}
