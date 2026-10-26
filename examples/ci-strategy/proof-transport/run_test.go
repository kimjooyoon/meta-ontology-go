package prooftransport

import (
	"context"
	"net/http"
	"sync"
	"testing"
)

func TestTransportAndIndependentReadback(t *testing.T) {
	for _, row := range []struct {
		name   string
		status int
		store  bool
		next   string
	}{
		{"success", 201, true, "ALREADY_PRESENT"},
		{"effect-despite-failed-response", 500, true, "ALREADY_PRESENT"},
		{"failed-with-no-effect", 500, false, "READY_TO_RESTORE"},
	} {
		t.Run(row.name, func(t *testing.T) {
			f := fixture(t, false)
			f.putStatus, f.store = row.status, row.store
			got := run(context.Background(), "transport-original-build-proof", f.material, "", true, f)
			if got.Decision.F_state != "READY_TO_RESTORE" || !got.Write.Attempted || got.Write.Status != int64(row.status) || got.Next == nil || got.Next.F_state != row.next || f.puts != 1 {
				t.Fatal(got, f.puts)
			}
			if got.Before.F_manifest_status != 404 || got.After == nil || got.Requests != 11 {
				t.Fatal("original observation or separate readback lost", got)
			}
		})
	}
}

func TestRepeatedAndConcurrentApplyKeepsOneContentEffect(t *testing.T) {
	f := fixture(t, false)
	var wait sync.WaitGroup
	for range 3 {
		wait.Go(func() {
			got := run(context.Background(), "transport-original-build-proof", f.material, "", true, f)
			if got.Decision.F_state != "READY_TO_RESTORE" && got.Decision.F_state != "ALREADY_PRESENT" {
				t.Error(got)
			}
		})
	}
	wait.Wait()
	if f.puts != 1 {
		t.Fatal("duplicate local write", f.puts)
	}
}

func TestUnchangedAndReadOnlyPlansDoNotWrite(t *testing.T) {
	for _, present := range []bool{false, true} {
		f := fixture(t, present)
		got := run(context.Background(), "transport-original-build-proof", f.material, "", false, f)
		if got.Write.Attempted || got.After != nil || f.puts != 0 {
			t.Fatal(got)
		}
	}
	f := fixture(t, false)
	got := run(context.Background(), "transport-original-build-proof", []byte("changed"), "", true, f)
	if got.Decision.F_state != "REFUTED_MATERIAL" || got.Requests != 0 || f.puts != 0 {
		t.Fatal(got)
	}
}

func TestCancelledBeforeAdmissionRetainsValidMaterial(t *testing.T) {
	f := fixture(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	executionSlot <- struct{}{}
	defer func() { <-executionSlot }()
	got := run(ctx, "transport-original-build-proof", f.material, "", true, f)
	if got.Decision.F_state != "REFRESH_OBSERVATION" || got.Write.Attempted || got.Requests != 0 {
		t.Fatal(got)
	}
}

func TestPostEffectCancellationRetainsAttemptAndUnknownReadback(t *testing.T) {
	f := fixture(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	transport := roundTrip(func(r *http.Request) (*http.Response, error) {
		res, err := f.RoundTrip(r)
		if r.Method == http.MethodPut {
			cancel()
		}
		return res, err
	})
	got := run(ctx, "transport-original-build-proof", f.material, "", true, transport)
	if !got.Write.Attempted || f.puts != 1 || !f.present || got.After == nil || got.After.F_read_error == "" {
		t.Fatal("lost attempted effect or unavailable readback", got)
	}
	if got.Next == nil || got.Next.F_state != "REFRESH_OBSERVATION" || got.Next.F_write_manifest {
		t.Fatal("incomplete post-effect observation became a retry", got)
	}
}
