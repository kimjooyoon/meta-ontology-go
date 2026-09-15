package sourceauthority

import "fmt"

const (
	Schema        = "gooo/source-backed-authority-contract/v1"
	MetricID      = "gooo.metric.semantic.source-backed-authority.v1"
	MetaOperation = "bind-source-backed-authority"
	ProofChoice   = "FOUNDATION"
	DenominatorID = "gooo/language-assurance-denominator/v1"
)

func Validate(contract Contract) error {
	if err := validateIdentity(contract); err != nil {
		return err
	}
	if err := validateMeasurement(contract); err != nil {
		return err
	}
	if err := validateSequences(contract); err != nil {
		return err
	}
	return validateFailureModes(contract)
}

func validateIdentity(contract Contract) error {
	values := []struct {
		name string
		got  string
		want string
	}{
		{"schema", contract.Schema, Schema},
		{"metric_id", contract.MetricID, MetricID},
		{"meta_operation", contract.MetaOperation, MetaOperation},
		{"proof_choice", contract.ProofChoice, ProofChoice},
		{"denominator_id", contract.DenominatorID, DenominatorID},
	}
	for _, value := range values {
		if value.got != value.want {
			return fmt.Errorf("%s: got %q want %q", value.name, value.got, value.want)
		}
	}
	return nil
}
