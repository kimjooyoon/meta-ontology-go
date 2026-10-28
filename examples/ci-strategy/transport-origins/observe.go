package transportorigins

import (
	"context"
	"encoding/json"

	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
	resolution "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/transport-resolution"
)

// ObserveForReassessment collects at most one source-selected read-only report.
// A nil report means the source did not request another observation.
func ObserveForReassessment(
	ctx context.Context, priorRaw, material []byte, token string,
) ([]byte, resolution.R_Resolution, error) {
	pending, err := ReassessSavedReports(priorRaw, nil)
	if err != nil || !pending.F_observation_needed {
		return nil, pending, err
	}
	report := transport.Run(ctx, pending.F_read_operation, material, token, false)
	raw, err := json.Marshal(report)
	if err != nil {
		return nil, pending, err
	}
	result, err := ReassessSavedReports(priorRaw, raw)
	return raw, result, err
}
