package prooftransport

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

// One process serializes write attempts. Registry content addresses, rather
// than this process-local slot, determine the identity across other writers.
var executionSlot = make(chan struct{}, 1)

type WriteObservation struct {
	Attempted bool   `json:"attempted"`
	Status    int64  `json:"status"`
	Error     string `json:"error,omitempty"`
}

type Report struct {
	ApplyRequested bool                    `json:"apply_requested"`
	Scope          R_TransportScope        `json:"scope"`
	Before         R_TransportObservation  `json:"before"`
	Decision       R_TransportStep         `json:"decision"`
	Write          WriteObservation        `json:"write"`
	After          *R_TransportObservation `json:"after,omitempty"`
	Next           *R_TransportStep        `json:"next,omitempty"`
	Requests       int64                   `json:"requests"`
	ResponseBytes  int64                   `json:"response_bytes"`
}

// Run consumes a registry bearer token owned by the calling runtime. Apply
// enables only the single source-selected manifest PUT; there are no retries.
func Run(ctx context.Context, operation string, material []byte, token string, apply bool) Report {
	return run(ctx, operation, material, token, apply, http.DefaultTransport)
}

func run(ctx context.Context, operation string, material []byte, token string, apply bool, transport http.RoundTripper) Report {
	scope := Plan(operation)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(scope.F_deadline_ms)*time.Millisecond)
	defer cancel()
	w := &wire{scope: scope, token: token, client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	result := Report{Scope: scope, ApplyRequested: apply}
	result.Before = R_TransportObservation{F_base_url: scope.F_base_url}
	if int64(len(material)) <= scope.F_max_response_bytes {
		material = bytes.Clone(material)
		result.Before.F_material_digest = digest(material)
	}
	result.Decision = Evaluate(operation, result.Before)
	if result.Decision.F_next_operation != "read-current-registry" {
		return result
	}
	if apply {
		select {
		case executionSlot <- struct{}{}:
			defer func() { <-executionSlot }()
		case <-ctx.Done():
			result.Before.F_read_error = "execution slot deadline"
			result.Decision = Evaluate(operation, result.Before)
			return result
		}
	}
	result.Before = w.observe(ctx, material)
	result.Decision = Evaluate(operation, result.Before)
	if apply && result.Decision.F_write_manifest {
		res, err := w.send(ctx, http.MethodPut, scope.F_base_url+"/manifests/"+result.Decision.F_manifest_digest, bytes.NewReader(material))
		result.Write.Attempted = res.attempted
		result.Write.Status = res.status
		if err != nil {
			result.Write.Error = err.Error()
		}
		after := w.observe(ctx, material)
		next := Evaluate(operation, after)
		result.After, result.Next = &after, &next
	}
	result.Requests, result.ResponseBytes = w.requests, w.bytes
	return result
}
