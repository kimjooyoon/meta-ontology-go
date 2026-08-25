package artifactemit

import (
	"encoding/json"
	"testing"
)

func TestEmitLowersUnknownPackageDecision(t *testing.T) {
	payload := validReceiptJSON(t)
	var value map[string]any
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatal(err)
	}
	value["decision"] = "UNRECOGNIZED"
	payload, _ = json.Marshal(value)
	artifact := Emit(OperationManifestKind, payload)
	if artifact.Decision != "FAIL_CLOSED" || artifact.Resolution != "LOWER_RESOLUTION" ||
		artifact.Reason != "PACKAGE_DECISION_UNKNOWN" {
		t.Fatalf("artifact=%#v", artifact)
	}
}

func TestEmitRegistryRejectsUnknownKind(t *testing.T) {
	artifact := Emit("not-registered", validReceiptJSON(t))
	if artifact.Decision != "FAIL_CLOSED" || artifact.Reason != "EMITTER_UNKNOWN" ||
		artifact.Extensions.RegisteredEmitters != 3 {
		t.Fatalf("artifact=%#v", artifact)
	}
}

func TestEmitProjectsSymbolicInvocationSchema(t *testing.T) {
	artifact := Emit(SymbolicInvocationSchemaKind, validReceiptJSON(t))
	schema := artifact.JSONSchema
	if artifact.Decision != "PASS" || artifact.Resolution != SymbolicInvocationResolution ||
		artifact.Schema != SymbolicInvocationArtifact || schema == nil ||
		schema.Dialect != JSONSchemaDraft202012 || schema.Properties.Activity.Const != "PayOrder" ||
		len(schema.Properties.Inputs.PrefixItems) != 1 ||
		schema.Properties.Inputs.PrefixItems[0].Const != "urn:order" ||
		schema.Properties.Inputs.MinItems != 1 || schema.Properties.Inputs.MaxItems != 1 {
		t.Fatalf("artifact=%#v", artifact)
	}
}

func validReceiptJSON(t *testing.T) []byte {
	t.Helper()
	value := packageReceipt{
		Schema: PackageReceiptSchema, Decision: "PASS", Resolution: "EXACT",
		PackagePath: "billing-package", Package: "billing", Namespace: "billing", Entry: "PayOrder",
		Sources: []sourceDefinition{{Filename: "billing.gooo", Digest: "sha256:source", DeclarationCount: 3}},
		Execution: executionReceipt{Entry: operationEntry{
			Package: "billing", Namespace: "billing", Activity: "PayOrder",
			Inputs: []Binding{{Name: "Order", ID: "urn:order"}}, Output: Binding{Name: "Receipt", ID: "urn:receipt"},
		}}, Digest: "sha256:receipt",
	}
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
