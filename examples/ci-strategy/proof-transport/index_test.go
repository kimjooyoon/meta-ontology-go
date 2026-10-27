package prooftransport

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReferrerIndexPresenceAndPagination(t *testing.T) {
	for _, row := range []struct {
		name, body, state string
		links             []string
	}{
		{"omitted-media-type", `{"schemaVersion":2,"manifests":[]}`, "READY_TO_RESTORE", nil},
		{"explicit-null", `{"schemaVersion":2,"mediaType":null,"manifests":[]}`, "REFRESH_OBSERVATION", nil},
		{"explicit-empty", `{"schemaVersion":2,"mediaType":"","manifests":[]}`, "REFRESH_OBSERVATION", nil},
		{"wrong-media-type", `{"schemaVersion":2,"mediaType":"text/plain","manifests":[]}`, "REFRESH_OBSERVATION", nil},
		{"missing-digest", `{"schemaVersion":2,"manifests":[{}]}`, "REFRESH_OBSERVATION", nil},
		{"later-link-header", `{"schemaVersion":2,"manifests":[]}`, "REFRESH_OBSERVATION", []string{"", `<https://registry.invalid/next>; rel="next"`}},
	} {
		t.Run(row.name, func(t *testing.T) {
			f := fixture(t, false)
			transport := roundTrip(func(r *http.Request) (*http.Response, error) {
				res, err := f.RoundTrip(r)
				if r.URL.String() == f.scope.F_referrers_url {
					res.Body = io.NopCloser(strings.NewReader(row.body))
					res.Header["Link"] = row.links
				}
				return res, err
			})
			got := run(context.Background(), "transport-original-build-proof", f.material, "", false, transport)
			if got.Decision.F_state != row.state || got.Write.Attempted || f.puts != 0 {
				t.Fatal("wrong completeness observation", got)
			}
		})
	}
}
