package executioncontinuation

import (
	"context"
	"net/http"
	"testing"
)

func TestMissingPopulationCountsFresh1621(t *testing.T) {
	for _, member := range []string{"artifacts", "workflow_runs"} {
		t.Run(member, func(t *testing.T) {
			client := Client{Repository: "kimjooyoon/meta-ontology-go", HTTP: &http.Client{
				Transport: transport1617(func(r *http.Request) (*http.Response, error) {
					return jsonResponse1619(r, map[string]any{member: []any{}}), nil
				}),
			}}
			var err error
			if member == "artifacts" {
				_, err = client.artifacts(context.Background(), 9162101)
			} else {
				_, err = client.HasReservation(context.Background(), Envelope{RequestDigest: Digest([]byte("1621"))})
			}
			state, _ := errorState(err)
			if err == nil || state != "UNKNOWN" {
				t.Fatal("absent total became empty complete population", err)
			}
		})
	}
}
