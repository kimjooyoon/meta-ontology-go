package prooftransport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
)

type registryFixture struct {
	mu                         sync.Mutex
	scope                      R_TransportScope
	material, subject, payload []byte
	present                    bool
	puts                       int
	putStatus                  int
	store                      bool
}

func fixture(t *testing.T, present bool) *registryFixture {
	t.Helper()
	read := func(name string) []byte {
		b, e := os.ReadFile("testdata/" + name + ".json")
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	f := &registryFixture{scope: Plan("transport-original-build-proof"), material: read("manifest"), subject: read("subject"), payload: read("payload"), present: present, putStatus: 201, store: true}
	if digest(f.material) != f.scope.F_manifest_digest || digest(f.subject) != f.scope.F_subject_digest || digest(f.payload) != f.scope.F_payload_digest {
		t.Fatal("original public bytes changed")
	}
	return f
}

func (f *registryFixture) RoundTrip(r *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	code, data, media := 404, []byte("missing"), "application/json"
	base := f.scope.F_base_url
	switch r.Method + " " + r.URL.String() {
	case "GET " + base + "/manifests/" + f.scope.F_subject_digest:
		code, data = 200, f.subject
	case "GET " + base + "/blobs/" + f.scope.F_payload_digest:
		code, data = 200, f.payload
	case "GET " + base + "/blobs/" + f.scope.F_config_digest:
		code, data = 200, []byte("{}")
	case "GET " + base + "/manifests/" + f.scope.F_manifest_digest:
		if f.present {
			code, data = 200, f.material
		}
	case "GET " + f.scope.F_referrers_url:
		items := []map[string]string{}
		if f.present {
			items = append(items, map[string]string{"digest": f.scope.F_manifest_digest})
		}
		media = "application/vnd.oci.image.index.v1+json"
		code = 200
		data, _ = json.Marshal(map[string]any{"schemaVersion": 2, "mediaType": media, "manifests": items})
	case "PUT " + base + "/manifests/" + f.scope.F_manifest_digest:
		f.puts++
		raw, _ := io.ReadAll(r.Body)
		if !bytes.Equal(raw, f.material) || r.Header.Get("Content-Type") != "application/vnd.oci.image.manifest.v1+json" {
			code = 400
			break
		}
		code = f.putStatus
		if f.store {
			f.present = true
		}
	}
	return &http.Response{StatusCode: code, Header: http.Header{"Content-Type": []string{media}}, Body: io.NopCloser(bytes.NewReader(data))}, nil
}
