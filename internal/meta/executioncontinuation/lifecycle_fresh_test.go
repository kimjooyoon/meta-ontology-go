package executioncontinuation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestProducerAndTargetLifecycleFresh1619(t *testing.T) {
	f := fixtureConsumer1617(t)
	for _, c := range []struct {
		name, state, reason string
		posts               int
	}{
		{"producer-failed", "CONTINUATION_REFUTED", "SOURCE_RUN_NOT_SUCCESSFUL", 0},
		{"producer-missing", "DISPATCH_UNKNOWN", "SOURCE_RUN_METADATA_UNKNOWN", 0},
		{"target-missing", "DISPATCH_SENT", "TARGET_RUN_METADATA_UNKNOWN", 1},
		{"target-conflict-and-missing", "CONTINUATION_REFUTED", "TARGET_RUN_MISMATCH", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			client := client1617(t, f, "")
			base, posts := client.HTTP.Transport, 0
			client.HTTP.Transport = transport1617(func(r *http.Request) (*http.Response, error) {
				if r.Method == http.MethodPost {
					posts++
				}
				var changed *Run
				producer := f.grantRun
				target := f.consumerRun
				if strings.HasSuffix(r.URL.Path, fmt.Sprintf("/runs/%d", producer.ID)) {
					if c.name == "producer-failed" {
						producer.Conclusion = "failure"
						changed = &producer
					}
					if c.name == "producer-missing" {
						producer.Path = ""
						changed = &producer
					}
				}
				if strings.HasSuffix(r.URL.Path, fmt.Sprintf("/runs/%d", target.ID)) {
					if c.name == "target-missing" {
						target.Path = ""
						changed = &target
					}
					if c.name == "target-conflict-and-missing" {
						target.Path = ""
						target.Attempt = 2
						changed = &target
					}
				}
				if changed == nil {
					return base.RoundTrip(r)
				}
				return jsonResponse1619(r, changed), nil
			})
			receipt, _ := client.Dispatch(context.Background(), f.prepared.Metadata.Source)
			if receipt.State != c.state || receipt.Reason != c.reason || posts != c.posts {
				t.Fatal("dispatch lifecycle", receipt, posts)
			}
		})
	}
}

func TestUnknownCausalityFresh1619(t *testing.T) {
	for _, c := range []struct{ reason, dependency, next, class string }{
		{"SOURCE_ARTIFACT_EXPIRED", "expired_source_artifact", "produce_fresh_source_artifact", "EXPIRED"},
		{"DISPATCH_RESPONSE_UNKNOWN", "dispatch_outcome", "observe_dispatch_outcome", "DIRECT_MISSING"},
		{"SUBJECT_SOURCE_UNKNOWN", "subject_blob", "observe_exact_subject_blob", "DIRECT_MISSING"},
	} {
		t.Run(c.reason, func(t *testing.T) {
			r := Diagnostic(Envelope{}, unknown(c.reason))
			u := r.Unknown
			if u == nil || u.Stage == "" || u.Reason != c.reason || u.Step != c.dependency ||
				u.Next != c.next || u.Class != c.class || len(u.BlockedBy) != 1 || u.BlockedBy[0] != c.dependency {
				t.Fatal("causal unknown incomplete", r)
			}
		})
	}
}

func TestRetainedReservationAcrossHeadsFresh1619(t *testing.T) {
	f := fixtureConsumer1617(t)
	for _, variant := range []string{"retained", "expired", "forged"} {
		t.Run(variant, func(t *testing.T) {
			client := client1617(t, f, "")
			run := run1617("consumer", 9161900, strings.Repeat("9", 40))
			reservation := reservationFor(f.prepared)
			reservation.ConsumerRunID = run.ID
			if variant == "forged" {
				reservation.SubjectSHA = strings.Repeat("8", 40)
			}
			reservation.Digest = ""
			reservation.Digest = digestValue(reservation)
			raw := zipJSON1617(t, map[string]any{"reservation.json": reservation})
			a := artifact1617(run, 9161901, ReservationName(f.prepared.Metadata.Source.RequestDigest), raw)
			a.Expired = variant == "expired"
			base := client.HTTP.Transport
			client.HTTP.Transport = transport1617(func(req *http.Request) (*http.Response, error) {
				switch {
				case strings.HasSuffix(req.URL.Path, "/self-improvement-execution-consumer.yml/runs"):
					if req.URL.Query().Get("head_sha") != "" {
						t.Fatal("history narrowed to one workflow head")
					}
					return jsonResponse1619(req, map[string]any{"total_count": 1, "workflow_runs": []Run{run}}), nil
				case strings.HasSuffix(req.URL.Path, fmt.Sprintf("/runs/%d/artifacts", run.ID)):
					return jsonResponse1619(req, map[string]any{"total_count": 1, "artifacts": []any{a}}), nil
				case strings.HasSuffix(req.URL.Path, fmt.Sprintf("/artifacts/%d/zip", a.ID)):
					return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(raw))}, nil
				default:
					return base.RoundTrip(req)
				}
			})
			exists, err := client.HasReservation(context.Background(), f.prepared.Metadata.Source)
			if variant == "retained" {
				if !exists || err != nil {
					t.Fatal("cross-head reservation lost", exists, err)
				}
			} else if err == nil || exists {
				t.Fatal("uncertain history treated as unused", variant, exists, err)
			}
		})
	}
}

func jsonResponse1619(r *http.Request, value any) *http.Response {
	raw, _ := json.Marshal(value)
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(raw)), Request: r}
}
