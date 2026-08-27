package languageproofartifactverifier

import "reflect"

type receiptBinding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type receiptEntry struct {
	Package   string           `json:"package"`
	Namespace string           `json:"namespace"`
	Activity  string           `json:"activity"`
	Inputs    []receiptBinding `json:"inputs"`
	Output    receiptBinding   `json:"output"`
}

type receiptEvent struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Subject  string `json:"subject"`
}

type receiptDiagnostic struct {
	Stage   string `json:"stage"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type receiptEffects struct {
	// These are source-execution receipt observations. They are deliberately
	// not named like the verifier's net repository delta or capability policy.
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type operationReceipt struct {
	Schema         string              `json:"schema"`
	Decision       string              `json:"decision"`
	Reason         string              `json:"reason"`
	Resolution     string              `json:"resolution"`
	Filename       string              `json:"filename"`
	SourceDigest   string              `json:"source_digest"`
	SemanticDigest string              `json:"semantic_digest,omitempty"`
	Entry          receiptEntry        `json:"entry"`
	Events         []receiptEvent      `json:"events"`
	Diagnostics    []receiptDiagnostic `json:"diagnostics"`
	Effects        receiptEffects      `json:"effects"`
	Digest         string              `json:"digest"`
}

func verifyOperation(receipt operationReceipt, sourceDigest, sourcePath string, want projection) bool {
	if receipt.Digest != receiptDigest(receipt) || receipt.Schema != "gooo/source-execution-receipt/v1" ||
		receipt.Decision != "PASS" || receipt.Reason != "SOURCE_ACTIVITY_EXECUTED" || receipt.Resolution != "EXACT" ||
		receipt.Filename != sourcePath || receipt.SourceDigest != sourceDigest || receipt.SemanticDigest != want.SemanticDigest || len(receipt.Diagnostics) != 0 ||
		receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return false
	}
	wantInputs := make([]receiptBinding, len(want.Inputs))
	for index, item := range want.Inputs {
		wantInputs[index] = receiptBinding{Name: item.Name, ID: item.ID}
	}
	wantEntry := receiptEntry{Package: want.Package, Namespace: want.Namespace, Activity: want.Activity,
		Inputs: wantInputs, Output: receiptBinding{Name: want.Output.Name, ID: want.Output.ID}}
	if !reflect.DeepEqual(receipt.Entry, wantEntry) || len(receipt.Events) != 4 {
		return false
	}
	wantEvents := []receiptEvent{{1, "SOURCE_PARSED", sourceDigest}, {2, "SEMANTIC_LOWERED", receipt.SemanticDigest},
		{3, "ACTIVITY_INVOKED", want.Activity}, {4, "ENTITY_PRODUCED", want.Output.ID}}
	return reflect.DeepEqual(receipt.Events, wantEvents)
}

func receiptDigest(receipt operationReceipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}
