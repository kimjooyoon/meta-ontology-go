package prooftransport

import (
	"context"
	"net/http"
	"testing"
)

func TestCancellationAfterPreflightRetainsUnsentWriteAndLaterObservation(t *testing.T) {
	f := fixture(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	transport := roundTrip(func(request *http.Request) (*http.Response, error) {
		response, err := f.RoundTrip(request)
		if request.URL.String() == f.scope.F_referrers_url {
			cancel()
		}
		return response, err
	})
	got := run(ctx, "transport-original-build-proof", f.material, "", true, transport)
	if got.Decision.F_state != "READY_TO_RESTORE" || got.Write.Attempted || got.Write.Error == "" || f.puts != 0 || got.After == nil || got.Next == nil {
		t.Fatal("fixture did not retain an unsent write and later observation", got)
	}
	if got.Outcome.F_state != "WRITE_NOT_SENT" || !got.Outcome.F_command_failed || got.Outcome.F_effect_observed {
		t.Fatal("a legitimate later observation was treated as a contradiction", got.Outcome)
	}
}
