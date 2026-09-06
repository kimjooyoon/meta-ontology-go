package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"runtime"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func DecodeRequest(raw []byte) (Request, error) {
	var request Request
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := uniqueJSON(raw); err != nil {
		return Request{}, err
	}
	if err := decoder.Decode(&request); err != nil {
		return Request{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Request{}, fmt.Errorf("registration request has trailing JSON")
	}
	return request, nil
}

// InspectInputs observes exact content, not semantic acceptance. Callers pin the
// returned digest in Request before Compile; no request is inferred from prose.
func InspectInputs(repository fs.FS, request Request) (string, string, error) {
	inputs, err := readInputs(repository, request)
	if err != nil {
		return "", "", err
	}
	return digestValue(inputs), digest(inputs[request.Case.Path]), nil
}

func readInputs(repository fs.FS, request Request) (map[string][]byte, error) {
	if request.BaseVersion < 22 || request.BaseVersion > 10000 ||
		!validPath(request.Case.Path) || !strings.HasPrefix(request.Case.Path, "examples/") ||
		!strings.HasSuffix(request.Case.Path, ".gooo") {
		return nil, failure("REFUTED", "validate-request", "REGISTRATION_INPUT_INVALID", "", "correct-explicit-input")
	}
	paths, err := sourceInputPaths(repository)
	if err != nil {
		return nil, err
	}
	paths = append(paths, corpusPath, denominatorPath(request.BaseVersion), request.Case.Path)
	history, err := fs.Glob(repository, closureRoot+"evidence/denominator*.json")
	if err != nil {
		return nil, err
	}
	paths = append(paths, history...)
	inputs := make(map[string][]byte)
	for _, path := range paths {
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			return nil, failure("UNKNOWN", "read-input:"+path, "REGISTRATION_INPUT_MISSING", "DIRECT_MISSING", "restore-input-snapshot")
		}
		inputs[path] = append([]byte(nil), raw...)
	}
	if _, err := fs.Stat(repository, denominatorPath(request.BaseVersion+1)); err == nil {
		return nil, failure("REFUTED", "append-denominator", "DENOMINATOR_VERSION_ALREADY_EXISTS", "", "select-current-baseline")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, failure("UNKNOWN", "inspect-next-version", "DENOMINATOR_AVAILABILITY_UNKNOWN", "DIRECT_MISSING", "restore-input-snapshot")
	}
	return inputs, nil
}

func Compile(repository fs.FS, request Request) (Plan, error) {
	inputs, err := readInputs(repository, request)
	if err != nil {
		return Plan{}, err
	}
	if request.SnapshotDigest != digestValue(inputs) || request.SourceDigest != digest(inputs[request.Case.Path]) ||
		request.Toolchain != runtime.Version() {
		return Plan{}, failure("UNKNOWN", "bind-input", "REGISTRATION_INPUT_STALE", "STALE", "refresh-exact-input-binding")
	}
	if err := recheckExecutionIdentity(request.ExecutionIdentity); err != nil {
		return Plan{}, err
	}
	if err := validateCase(repository, request, inputs); err != nil {
		return Plan{}, err
	}
	binding, err := bindContract()
	if err != nil {
		return Plan{}, err
	}
	return Plan{request: request, inputs: inputs, digest: digestValue(inputs), binding: binding}, nil
}

func validateCase(repository fs.FS, request Request, inputs map[string][]byte) error {
	definition := request.Case
	if definition.ID == "" || definition.Kind != languagesyntax.KindValid ||
		definition.ExpectedDecision != languagesyntax.DecisionPass || definition.ExpectedDiagnostic != "" ||
		definition.Scope != languagesyntax.ScopeLanguageCapability || definition.MetaOperation != "replay-language-syntax" ||
		(definition.ProofChoice != "FOUNDATION" && definition.ProofChoice != "COHERENCE" && definition.ProofChoice != "REGRESSION") ||
		(definition.EntityFields && definition.ImplicitActivityPorts) {
		return failure("REFUTED", "validate-case", "REGISTRATION_CASE_CONTRACT_MISMATCH", "", "correct-explicit-case")
	}
	var registry languagesyntax.Registry
	if err := decodeStrict(inputs[corpusPath], &registry); err != nil {
		return err
	}
	for _, existing := range registry.Cases {
		if existing.ID == definition.ID || existing.Path == definition.Path {
			return failure("REFUTED", "register-case", "REGISTRATION_CASE_ALREADY_EXISTS", "", "report-counterexample")
		}
	}
	if slices.Contains(registry.MetaSources, definition.Path) {
		return failure("REFUTED", "register-case", "REGISTRATION_SOURCE_ALREADY_REGISTERED", "", "report-counterexample")
	}
	for _, unit := range registry.PackageUnits {
		if slices.Contains(unit.Members, definition.Path) {
			return failure("REFUTED", "register-case", "REGISTRATION_SOURCE_ALREADY_REGISTERED", "", "report-counterexample")
		}
	}
	observed := replay.Execute(repository, definition.Path, definition.Kind, "")
	if definition.EntityFields {
		observed = replay.ExecuteWithEntityFieldsSupport(repository, definition.Path, definition.Kind, "")
	} else if definition.ImplicitActivityPorts {
		observed = replay.ExecuteWithImplicitActivityPorts(repository, definition.Path, definition.Kind, "")
	}
	if !observed.ASTReplayed || !observed.ByteReplayed || !observed.SemanticReplayed || !observed.GetPut || !observed.PutGet {
		return failure("REFUTED", "replay-source", "REGISTRATION_SOURCE_CONFORMANCE_FAILED", "", "report-source-counterexample")
	}
	return nil
}
