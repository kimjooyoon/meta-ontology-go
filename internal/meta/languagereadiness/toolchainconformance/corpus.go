package toolchainconformance

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var fixedSurfaces = []SurfaceDefinition{
	{"language-syntax-roundtrip", "gooo/language-syntax-roundtrip/v1", 29, 16, 3, []string{"examples/ambiguity-budget/main.gooo"}},
	{"language-semantic-model", "gooo/language-semantic-model/v1", 31, 19, 3, []string{"examples/ambiguity-budget/main.gooo"}},
	{"language-deterministic-query", "gooo/language-deterministic-query/v1", 32, 18, 3, nil},
	{"language-go-interoperation", "gooo/language-go-interoperation/v1", 24, 18, 3, nil},
	{"language-diagnostic-provenance", "gooo/language-diagnostic-provenance/v1", 18, 18, 3, nil},
	{"language-package-runtime", "gooo/language-package-runtime-report/v1", 18, 18, 3, nil},
	{"toolchain-cli", "gooo/toolchain-cli-report/v2", 12, 18, 3, nil},
	{"toolchain-format-fix", "gooo/toolchain-format-fix-report/v1", 12, 18, 3, nil},
	{"toolchain-executable-use-cases", "gooo/toolchain-executable-use-cases/v1", 3, 8, 3, nil},
}

var fixedTamperCases = []TamperDefinition{
	{"missing-surface", "MISSING_SURFACE", "language-syntax-roundtrip"},
	{"unexpected-surface", "UNEXPECTED_SURFACE", "language-syntax-roundtrip"},
	{"schema-drift", "SCHEMA", "language-semantic-model"},
	{"head-drift", "HEAD", "language-deterministic-query"},
	{"decision-unknown", "DECISION", "language-go-interoperation"},
	{"resolution-descent", "RESOLUTION", "language-diagnostic-provenance"},
	{"unresolved-evidence", "UNRESOLVED", "language-package-runtime"},
	{"case-mismatch", "CASE", "toolchain-cli"},
	{"indicator-failure", "INDICATOR", "toolchain-format-fix"},
	{"proof-failure", "PROOF", "toolchain-executable-use-cases"},
	{"digest-failure", "DIGEST", "language-syntax-roundtrip"},
	{"repository-write", "WRITE", "toolchain-format-fix"},
	{"mutation-authority", "MUTATION", "toolchain-format-fix"},
}

func parseCorpus(raw []byte) (Corpus, string, error) {
	corpus := Corpus{}
	if err := json.Unmarshal(raw, &corpus); err != nil {
		return corpus, "", fmt.Errorf("FAIL_CLOSED: decode conformance corpus: %w", err)
	}
	if corpus.Schema != CorpusSchema ||
		!reflect.DeepEqual(corpus.Surfaces, fixedSurfaces) ||
		!reflect.DeepEqual(corpus.TamperCases, fixedTamperCases) {
		return corpus, "", fmt.Errorf("FAIL_CLOSED: conformance corpus drift")
	}
	return corpus, digestValue(corpus), nil
}
