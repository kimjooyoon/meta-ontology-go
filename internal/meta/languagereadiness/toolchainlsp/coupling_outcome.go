package toolchainlsp

import "github.com/kimjooyoon/meta-ontology-go/internal/lsp/coupling"

func couplingOutcome(envelope coupling.Envelope, status coupling.Outcome, reason coupling.Reason, request coupling.Request) (coupling.Result, error) {
	envelope.Status, envelope.Reason = status, reason
	raw, err := couplingBytes(envelope)
	if err != nil {
		return coupling.Result{}, err
	}
	adapter, err := coupling.New(raw)
	if err != nil {
		return coupling.Result{}, err
	}
	return adapter.Resolve(request), nil
}
