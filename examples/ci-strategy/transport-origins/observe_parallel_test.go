package transportorigins

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

type overlappingReportReads struct {
	registry *reportRegistry
	mu       sync.Mutex
	entered  chan struct{}
	release  chan struct{}
}

func (gate *overlappingReportReads) RoundTrip(request *http.Request) (*http.Response, error) {
	subject := gate.registry.scope.F_base_url + "/manifests/" + gate.registry.scope.F_subject_digest
	if request.Method == http.MethodGet && request.URL.String() == subject {
		select {
		case gate.entered <- struct{}{}:
		case <-request.Context().Done():
			return nil, request.Context().Err()
		}
		select {
		case <-gate.release:
		case <-request.Context().Done():
			return nil, request.Context().Err()
		}
	}
	gate.mu.Lock()
	defer gate.mu.Unlock()
	return gate.registry.RoundTrip(request)
}

func TestObserveForReassessmentCanOverlapReadOnlyCollections(t *testing.T) {
	_, raw, wire := runSavedReport(t, true, 500, true, true)
	wire.cancel = nil
	puts := wire.puts
	gate := &overlappingReportReads{registry: wire, entered: make(chan struct{}, 2), release: make(chan struct{})}
	http.DefaultTransport = gate
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type outcome struct {
		state string
		err   error
	}
	results := make(chan outcome, 2)
	for range 2 {
		go func() {
			_, result, err := ObserveForReassessment(ctx, raw, wire.material, "")
			results <- outcome{result.F_state, err}
		}()
	}
	overlapped := true
	for range 2 {
		select {
		case <-gate.entered:
		case <-ctx.Done():
			overlapped = false
		}
	}
	close(gate.release)
	for range 2 {
		select {
		case result := <-results:
			if result.err != nil || result.state != "READ_REPORT_EFFECT_OBSERVED" {
				t.Error("independent read did not return its assessment", result)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("read-only collection did not terminate")
		}
	}
	if !overlapped || wire.puts != puts {
		t.Fatal("read-only collections serialized or performed another write", overlapped, wire.puts-puts)
	}
}
