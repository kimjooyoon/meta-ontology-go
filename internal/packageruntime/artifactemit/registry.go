package artifactemit

import (
	"encoding/json"
	"sort"
	"strings"
)

type emitter struct {
	kind    string
	project func(packageReceipt) Artifact
}

func emitterRegistry() []emitter {
	return []emitter{
		{kind: OperationManifestKind, project: projectOperationManifest},
		{kind: OperationInterfaceKind, project: projectOperationInterface},
		{kind: SymbolicInvocationSchemaKind, project: projectSymbolicInvocationSchema},
	}
}

func RegisteredKinds() []string {
	registeredEmitters := emitterRegistry()
	kinds := make([]string, len(registeredEmitters))
	for index, registered := range registeredEmitters {
		kinds[index] = registered.kind
	}
	sort.Strings(kinds)
	return kinds
}

func Emit(kind string, payload []byte) Artifact {
	var source packageReceipt
	if err := json.Unmarshal(payload, &source); err != nil {
		return failed(kind, "EXACT", "PACKAGE_RECEIPT_INVALID")
	}
	if source.Schema != PackageReceiptSchema {
		return failed(kind, "EXACT", "PACKAGE_RECEIPT_SCHEMA_UNKNOWN")
	}
	if source.Decision != "PASS" {
		return failed(kind, "LOWER_RESOLUTION", "PACKAGE_DECISION_UNKNOWN")
	}
	if !receiptComplete(source) {
		return failed(kind, "EXACT", "PACKAGE_RECEIPT_INCOMPLETE")
	}
	if source.Effects.RepositoryWrites != 0 || source.Effects.MutationAuthority {
		return failed(kind, "EXACT", "PACKAGE_EFFECTS_OBSERVED")
	}
	for _, registered := range emitterRegistry() {
		if registered.kind == kind {
			return registered.project(source)
		}
	}
	return failed(kind, "LOWER_RESOLUTION", "EMITTER_UNKNOWN")
}

func receiptComplete(source packageReceipt) bool {
	entry := source.Execution.Entry
	if source.PackagePath == "" || source.Package == "" || source.Namespace == "" ||
		source.Entry != entry.Activity || entry.Package != source.Package || entry.Namespace != source.Namespace ||
		entry.Activity == "" || entry.Output.Name == "" || entry.Output.ID == "" || len(source.Sources) == 0 ||
		!strings.HasPrefix(source.Digest, "sha256:") {
		return false
	}
	for _, definition := range source.Sources {
		if definition.Filename == "" || !strings.HasPrefix(definition.Digest, "sha256:") || definition.DeclarationCount < 1 {
			return false
		}
	}
	for _, input := range entry.Inputs {
		if input.Name == "" || input.ID == "" {
			return false
		}
	}
	return true
}
