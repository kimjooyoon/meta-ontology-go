package decisionroute

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func validTestRequest() Request {
	return Request{
		Schema: RequestSchema, State: "state",
		Question: Question{
			ID: "decision", Instructions: "Choose an option.",
			Options: []Option{{ID: "yes", Description: "Proceed"}, {ID: "no", Description: "Stop"}},
		},
		Fallback: "no",
	}
}

func TestProviderModelDigestBindingAndEmptyCompatibility(t *testing.T) {
	request := validTestRequest()
	got, err := Validate(request)
	if err != nil {
		t.Fatal(err)
	}
	type legacyRequest struct {
		Schema   string   `json:"schema"`
		State    string   `json:"state"`
		Question Question `json:"question"`
		Fallback string   `json:"fallback"`
	}
	legacyJSON, err := json.Marshal(legacyRequest{
		Schema: request.Schema, State: request.State, Question: request.Question, Fallback: request.Fallback,
	})
	if err != nil {
		t.Fatal(err)
	}
	legacyDigest := sha256.Sum256(legacyJSON)
	want := "sha256:" + hex.EncodeToString(legacyDigest[:])
	if got != want {
		t.Fatalf("empty provider model changed the historical digest: got %s, want %s", got, want)
	}

	request.ProviderModel = "english"
	englishDigest, err := Validate(request)
	if err != nil {
		t.Fatal(err)
	}
	request.ProviderModel = "multilingual"
	multilingualDigest, err := Validate(request)
	if err != nil {
		t.Fatal(err)
	}
	if englishDigest == got || multilingualDigest == got || englishDigest == multilingualDigest {
		t.Fatalf("explicit provider model was not bound into the digest: empty=%s english=%s multilingual=%s", got, englishDigest, multilingualDigest)
	}
}

func TestValidateRejectsUnsupportedProviderModel(t *testing.T) {
	for _, model := range []string{"English", "typed-decision", "unknown"} {
		t.Run(model, func(t *testing.T) {
			request := validTestRequest()
			request.ProviderModel = model
			if _, err := Validate(request); err == nil {
				t.Fatalf("Validate accepted unsupported provider model %q", model)
			}
		})
	}
}

func TestValidateProviderModelAcceptsSupportedValues(t *testing.T) {
	for _, model := range []string{"", "english", "multilingual", "typed-decisions"} {
		if err := ValidateProviderModel(model); err != nil {
			t.Errorf("ValidateProviderModel(%q) returned %v", model, err)
		}
	}
}

func TestResolveSendsExplicitProviderModelAndRecordsReceipt(t *testing.T) {
	request := validTestRequest()
	request.ProviderModel = "typed-decisions"
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode Laya request: %v", err)
		}
		writeLayaResult(w, `"model":"typed-decisions"`)
	}))
	defer server.Close()

	receipt, err := Resolve(context.Background(), request, server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if received["model"] != request.ProviderModel {
		t.Fatalf("Laya request model = %v, want %q", received["model"], request.ProviderModel)
	}
	if receipt.Mode != "laya" || receipt.Selected != "yes" || receipt.RequestedProviderModel != request.ProviderModel {
		t.Fatalf("unexpected Laya receipt: %+v", receipt)
	}
}

func TestResolveEmptyProviderModelPreservesAutomaticRequest(t *testing.T) {
	request := validTestRequest()
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode Laya request: %v", err)
		}
		writeLayaResult(w, "")
	}))
	defer server.Close()

	receipt, err := Resolve(context.Background(), request, server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := received["model"]; exists {
		t.Fatalf("automatic Laya request unexpectedly included a model: %v", received["model"])
	}
	if receipt.Mode != "laya" || receipt.RequestedProviderModel != "" {
		t.Fatalf("unexpected automatic receipt: %+v", receipt)
	}
}

func TestResolveRejectsExplicitProviderModelMismatchOrMissingRouteModel(t *testing.T) {
	for _, tc := range []struct {
		name        string
		routingJSON string
	}{
		{name: "mismatch", routingJSON: `"model":"multilingual"`},
		{name: "missing", routingJSON: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := validTestRequest()
			request.ProviderModel = "english"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeLayaResult(w, tc.routingJSON)
			}))
			defer server.Close()

			receipt, err := Resolve(context.Background(), request, server.URL, "")
			if err != nil {
				t.Fatal(err)
			}
			if receipt.Mode != "deterministic_fallback" || receipt.FallbackReason != FallbackInvalidResult ||
				receipt.Selected != request.Fallback || receipt.RequestedProviderModel != request.ProviderModel {
				t.Fatalf("unexpected mismatch fallback: %+v", receipt)
			}
		})
	}
}

func TestResolveAbsentEndpointRetainsDeclaredFallbackAndModel(t *testing.T) {
	request := validTestRequest()
	request.ProviderModel = "multilingual"
	receipt, err := Resolve(context.Background(), request, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Mode != "deterministic_fallback" || receipt.FallbackReason != FallbackNotConfigured ||
		receipt.Selected != request.Fallback || receipt.RequestedProviderModel != request.ProviderModel {
		t.Fatalf("unexpected absent-endpoint receipt: %+v", receipt)
	}
}

func writeLayaResult(w http.ResponseWriter, routingModel string) {
	w.Header().Set("Content-Type", "application/json")
	routing := ""
	if routingModel != "" {
		routing = routingModel
	}
	_, _ = w.Write([]byte(`{"model":"backend-model","routing":{` + routing + `},"answers":{"decision":{"choice":"yes"}}}`))
}
