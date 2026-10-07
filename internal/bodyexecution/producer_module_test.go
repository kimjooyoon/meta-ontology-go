package bodyexecution

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/buildidentity"
)

func TestNativeObservationModuleWireFormat(t *testing.T) {
	module := &buildidentity.Module{Path: "example.org/compiler", Version: "v1.2.3", Sum: "h1:observed"}
	values := []any{
		Observation{ProducerSourceSHA: "UNBOUND_LOCAL_SOURCE", ProducerModule: module},
		CompositionRuntime{ProducerSourceSHA: "UNBOUND_LOCAL_SOURCE", ProducerModule: module},
	}
	for _, value := range values {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var got struct {
			Module *buildidentity.Module `json:"producer_module"`
			Source string                `json:"producer_source_sha"`
		}
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got.Module, module) || got.Source != "UNBOUND_LOCAL_SOURCE" {
			t.Fatalf("module observation = %s", raw)
		}
	}
}

func TestInitialNativeObservationsUseCurrentModule(t *testing.T) {
	want := buildidentity.MainModule()
	single := initialResult(nil, bodycodegen.Result{}, nil, nil)
	graph := initialCompositionRuntime(nil, Composition{}, CompositionCases{})
	if !reflect.DeepEqual(single.Observation.ProducerModule, want) || !reflect.DeepEqual(graph.ProducerModule, want) {
		t.Fatalf("current module missing from native initial observations")
	}
	module := &buildidentity.Module{Path: "example.org/compiler", Version: "v1.2.3", Sum: "h1:observed"}
	single.Observation.ProducerModule = module
	receipt := runtimeCompleteness(bodycodegen.Result{}, single)
	if !reflect.DeepEqual(receipt.Scope["producer_module"], module) {
		t.Fatalf("module missing from runtime receipt scope")
	}
}

func TestNativeObservationAcceptsOlderReceiptsWithoutModule(t *testing.T) {
	for _, value := range []any{&Observation{}, &CompositionRuntime{}} {
		if err := json.Unmarshal([]byte(`{"producer_source_sha":"UNBOUND_LOCAL_SOURCE"}`), value); err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), `"producer_module"`) {
			t.Fatalf("invented module for older receipt: %s", raw)
		}
	}
}

func TestRuntimeReceiptDoesNotBorrowPriorModule(t *testing.T) {
	prior := bodycodegen.Result{Report: bodycodegen.Report{
		CompletenessReceipt: bodycodegen.FailureCompletenessReceipt("Example", nil, "unobserved"),
	}}
	prior.Report.CompletenessReceipt.Scope["producer_module"] = &buildidentity.Module{
		Path: "example.org/previous-compiler", Version: "v1.0.0",
	}
	parent, err := json.Marshal(prior.Report.CompletenessReceipt)
	if err != nil {
		t.Fatal(err)
	}
	got := runtimeCompleteness(prior, Result{ParentReceipt: parent})
	if _, retained := got.Scope["producer_module"]; retained {
		t.Fatal("unobserved current module borrowed the prior producer's identity")
	}
}
