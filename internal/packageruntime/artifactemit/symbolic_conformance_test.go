package artifactemit

import (
	"encoding/json"
	"testing"
)

const symbolicConformanceReceipt = `{"schema":"gooo/package-source-execution-receipt/v1","decision":"PASS","resolution":"EXACT","package_path":"checkout","package":"checkout","namespace":"checkout","entry":"Checkout","sources":[{"filename":"activity.gooo","digest":"sha256:source","declaration_count":1}],"execution":{"entry":{"package":"checkout","namespace":"checkout","activity":"Checkout","inputs":[{"name":"Cart","id":"urn:gooo:checkout:cart"},{"name":"PaymentMethod","id":"urn:gooo:checkout:payment-method"}],"output":{"name":"Receipt","id":"urn:gooo:checkout:receipt"}}},"effects":{"repository_writes":0,"mutation_authority":false},"digest":"sha256:source"}`

func TestSymbolicConformanceVectorsAreCompilerProjected(t *testing.T) {
	artifact := Emit(SymbolicInvocationSchemaKind, []byte(symbolicConformanceReceipt))
	payload, err := Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Conformance SymbolicInvocationConformance `json:"conformance"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	conformance := decoded.Conformance
	if conformance.Schema != SymbolicInvocationConformanceSchema ||
		conformance.Decision != "PASS" || conformance.Resolution != "STRUCTURAL_ONLY" {
		t.Fatalf("unexpected conformance identity: %+v", conformance)
	}
	if conformance.GeneratedVectors != 2 || conformance.EmbeddedHandwrittenVectors != 0 ||
		len(conformance.Vectors) != 2 {
		t.Fatalf("unexpected vector counts: %+v", conformance)
	}
	accepted, rejected := conformance.Vectors[0], conformance.Vectors[1]
	if accepted.Expected != "ACCEPT" || accepted.ProofChoice != "FOUNDATION" ||
		accepted.Instance.Activity != "Checkout" || len(accepted.Instance.Inputs) != 2 {
		t.Fatalf("unexpected accepted vector: %+v", accepted)
	}
	if rejected.Expected != "REJECT" || rejected.ProofChoice != "REGRESSION" ||
		rejected.Instance.Activity != "" || len(rejected.Instance.Inputs) != 2 {
		t.Fatalf("unexpected rejected vector: %+v", rejected)
	}
	if conformance.Effects.RepositoryWrites != 0 || conformance.Effects.MutationAuthority {
		t.Fatalf("unexpected conformance effects: %+v", conformance.Effects)
	}
	if !ValidDigest(artifact) {
		t.Fatal("artifact digest does not cover compiler-projected conformance")
	}
}
