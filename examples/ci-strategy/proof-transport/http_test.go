package prooftransport

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testWire(transport http.RoundTripper) *wire {
	return &wire{scope: Plan("transport-original-build-proof"), token: "private-test-token", client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func TestRedirectRetainsBudgetsAndDropsCrossHostCredentials(t *testing.T) {
	calls := 0
	w := testWire(roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			if r.Header.Get("Authorization") == "" {
				t.Fatal("registry token absent")
			}
			return &http.Response{StatusCode: 307, Header: http.Header{"Location": []string{"https://blob.invalid/content"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
		}
		if r.Header.Get("Authorization") != "" {
			t.Fatal("credential forwarded to another host")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}"))}, nil
	}))
	res, err := w.read(context.Background(), w.scope.F_base_url+"/blobs/any")
	if err != nil || res.status != 200 || w.requests != 2 || w.bytes != 2 {
		t.Fatal(res, err, w.requests, w.bytes)
	}
	w.scope.F_max_requests = 2
	if _, err = w.read(context.Background(), w.scope.F_base_url); err == nil || calls != 2 {
		t.Fatal("request ceiling ignored", err, calls)
	}
}

func TestBodyBudgetCannotAuthorizeOrHidePartialIndex(t *testing.T) {
	w := testWire(roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 100)))}, nil
	}))
	w.scope.F_max_total_bytes = 8
	if _, err := w.read(context.Background(), w.scope.F_base_url); err == nil || w.bytes != 9 {
		t.Fatal("bounded sentinel read", err, w.bytes)
	}
	if res, err := w.send(context.Background(), http.MethodPut, w.scope.F_base_url, strings.NewReader("manifest")); err == nil || res.attempted {
		t.Fatal("exhausted budget attempted a write", res, err)
	}
	f := fixture(t, false)
	for _, body := range []string{`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json"}`, `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":null}`} {
		transport := roundTrip(func(r *http.Request) (*http.Response, error) {
			res, err := f.RoundTrip(r)
			if r.URL.String() == f.scope.F_referrers_url {
				res.Body = io.NopCloser(strings.NewReader(body))
			}
			return res, err
		})
		got := run(context.Background(), "transport-original-build-proof", f.material, "", true, transport)
		if got.Decision.F_state != "REFRESH_OBSERVATION" || f.puts != 0 {
			t.Fatal("absent list became complete empty list", got)
		}
	}
}
