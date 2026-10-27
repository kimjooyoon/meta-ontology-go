package transportorigins

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
)

type reportRegistry struct {
	scope                      transport.R_TransportScope
	material, subject, payload []byte
	present, store             bool
	status, puts, requests     int
	cancel                     context.CancelFunc
}

func runSavedReport(t *testing.T, apply bool, status int, store, cancelReadback bool) (transport.Report, []byte, *reportRegistry) {
	t.Helper()
	read := func(name string) []byte {
		raw, err := os.ReadFile("../proof-transport/testdata/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	wire := &reportRegistry{scope: transport.Plan("transport-original-build-proof"), material: read("manifest"), subject: read("subject"), payload: read("payload"), status: status, store: store}
	previous := http.DefaultTransport
	http.DefaultTransport = wire
	t.Cleanup(func() { http.DefaultTransport = previous })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if cancelReadback {
		wire.cancel = cancel
	}
	report := transport.Run(ctx, "transport-original-build-proof", wire.material, "", apply)
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return report, raw, wire
}

func (wire *reportRegistry) RoundTrip(request *http.Request) (*http.Response, error) {
	wire.requests++
	code, data, media := 404, []byte("missing"), "application/json"
	switch request.Method + " " + request.URL.String() {
	case "GET " + wire.scope.F_base_url + "/manifests/" + wire.scope.F_subject_digest:
		code, data = 200, wire.subject
	case "GET " + wire.scope.F_base_url + "/blobs/" + wire.scope.F_payload_digest:
		code, data = 200, wire.payload
	case "GET " + wire.scope.F_base_url + "/blobs/" + wire.scope.F_config_digest:
		code, data = 200, []byte("{}")
	case "GET " + wire.scope.F_base_url + "/manifests/" + wire.scope.F_manifest_digest:
		if wire.present {
			code, data = 200, wire.material
		}
	case "GET " + wire.scope.F_referrers_url:
		entries := []map[string]string{}
		if wire.present {
			entries = append(entries, map[string]string{"digest": wire.scope.F_manifest_digest})
		}
		media = "application/vnd.oci.image.index.v1+json"
		code = 200
		data, _ = json.Marshal(map[string]any{"schemaVersion": 2, "mediaType": media, "manifests": entries})
	case "PUT " + wire.scope.F_base_url + "/manifests/" + wire.scope.F_manifest_digest:
		wire.puts++
		code = wire.status
		if wire.store {
			wire.present = true
		}
		if wire.cancel != nil {
			wire.cancel()
		}
	}
	return &http.Response{StatusCode: code, Header: http.Header{"Content-Type": []string{media}}, Body: io.NopCloser(bytes.NewReader(data))}, nil
}
