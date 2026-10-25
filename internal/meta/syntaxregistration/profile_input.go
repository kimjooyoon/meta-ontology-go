package syntaxregistration

import (
	"bytes"
	"fmt"
	"io/fs"
	"reflect"
	"runtime"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func DecodeProfileMigrationRequest(raw []byte) (ProfileMigrationRequest, error) {
	var request ProfileMigrationRequest
	err := decodeStrict(raw, &request)
	return request, err
}

func readProfileInputs(repository fs.FS, request ProfileMigrationRequest) (map[string][]byte, error) {
	inputs, err := readInputs(repository, Request{Case: request.Case, BaseVersion: request.BaseVersion})
	if err != nil {
		return nil, err
	}
	raw, err := fs.ReadFile(repository, profileContractPath)
	if err != nil {
		return nil, profileIssue("UNKNOWN", "read-contract", "PROFILE_CONTRACT_SOURCE_MISSING")
	}
	if !bytes.Equal(raw, profileMigrationSource) {
		return nil, profileIssue("REFUTED", "bind-contract", "PROFILE_CONTRACT_SOURCE_DIVERGED")
	}
	inputs[profileContractPath] = bytes.Clone(raw)
	if err := readProfileFamilyInputs(repository, inputs, request.Case); err != nil {
		return nil, err
	}
	return inputs, nil
}

// InspectProfileMigrationInputs observes bytes only; it grants no source writes.
func InspectProfileMigrationInputs(repository fs.FS, request ProfileMigrationRequest,
	replacement []byte) (string, string, string, error) {
	inputs, err := readProfileInputs(repository, request)
	if err != nil {
		return "", "", "", err
	}
	return digestValue(inputs), digest(inputs[request.Case.Path]), digest(replacement), nil
}

func CompileProfileMigration(repository fs.FS, request ProfileMigrationRequest,
	replacement []byte) (ProfileMigrationPlan, error) {
	if err := validateProfileMigrationRequest(request, replacement); err != nil {
		return ProfileMigrationPlan{}, err
	}
	inputs, err := readProfileInputs(repository, request)
	if err != nil {
		return ProfileMigrationPlan{}, err
	}
	if request.SnapshotDigest != digestValue(inputs) || request.SourceDigest != digest(inputs[request.Case.Path]) ||
		request.ReplacementDigest != digest(replacement) || request.Toolchain != runtime.Version() {
		return ProfileMigrationPlan{}, profileIssue("UNKNOWN", "pin-input", "PROFILE_MIGRATION_INPUT_STALE")
	}
	if err := recheckExecutionIdentity(request.ExecutionIdentity); err != nil {
		return ProfileMigrationPlan{}, err
	}
	binding, err := bindProfileContract(inputs[profileContractPath])
	if err != nil {
		return ProfileMigrationPlan{}, err
	}
	plan := ProfileMigrationPlan{request: request, inputs: inputs,
		replacement: bytes.Clone(replacement), binding: binding}
	if err := plan.bindProfileCase(); err != nil {
		return ProfileMigrationPlan{}, err
	}
	if err := plan.replayProfileSources(); err != nil {
		return ProfileMigrationPlan{}, err
	}
	return plan, nil
}

func validateProfileMigrationRequest(request ProfileMigrationRequest, replacement []byte) error {
	definition := request.Case
	if request.Schema != ProfileMigrationSchema || len(replacement) == 0 || len(replacement) > profileSourceLimit ||
		definition.Path == profileContractPath ||
		definition.ID == "" || definition.Kind != languagesyntax.KindValid ||
		definition.ExpectedDecision != languagesyntax.DecisionPass || definition.ExpectedDiagnostic != "" ||
		(definition.ProofChoice != "FOUNDATION" && definition.ProofChoice != "COHERENCE" &&
			definition.ProofChoice != "REGRESSION") ||
		definition.Scope != languagesyntax.ScopeLanguageCapability || definition.MetaOperation != "replay-language-syntax" ||
		(definition.EntityFields && definition.ImplicitActivityPorts) ||
		(definition.OrderedCollections && (definition.EntityFields || definition.ImplicitActivityPorts)) ||
		!languagesyntax.ValidProfileSelection(definition) {
		return profileIssue("REFUTED", "validate-request", "PROFILE_MIGRATION_REQUEST_INVALID")
	}
	if _, known := replay.EntityFieldsProfileSupport(request.TargetProfile); !known {
		return profileIssue("REFUTED", "select-profile", "PROFILE_TARGET_UNSUPPORTED")
	}
	if reflect.DeepEqual(definition, profileAfterCase(request)) {
		return profileIssue("REFUTED", "select-profile", "PROFILE_SELECTION_UNCHANGED")
	}
	return nil
}

func profileAfterCase(request ProfileMigrationRequest) languagesyntax.CaseDefinition {
	after := request.Case
	after.EntityFields, after.ImplicitActivityPorts, after.OrderedCollections = false, false, false
	after.EntityFieldsProfile = request.TargetProfile
	return after
}

func (plan *ProfileMigrationPlan) bindProfileCase() error {
	if err := decodeStrict(plan.inputs[corpusPath], &plan.registry); err != nil {
		return err
	}
	if plan.registry.Schema != languagesyntax.RegistrySchema {
		return profileIssue("REFUTED", "bind-corpus", "PROFILE_CORPUS_SCHEMA_MISMATCH")
	}
	found := -1
	for index, definition := range plan.registry.Cases {
		if definition.ID != plan.request.Case.ID && definition.Path != plan.request.Case.Path {
			continue
		}
		if found >= 0 || !reflect.DeepEqual(definition, plan.request.Case) {
			return profileIssue("REFUTED", "bind-case", "PROFILE_CASE_IDENTITY_CONTRADICTION")
		}
		found = index
	}
	if found < 0 {
		return profileIssue("UNKNOWN", "bind-case", "PROFILE_CASE_NOT_REGISTERED")
	}
	plan.index = found
	native, err := parseSourceUnits(plan.inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		return err
	}
	if err := bindProfileMetaSources(native, plan.registry.MetaSources); err != nil {
		return err
	}
	if err := bindNativeCaseSources(native, plan.registry.CaseSources); err != nil {
		return err
	}
	rows, err := readProfileNativeCases(native)
	if err != nil {
		return err
	}
	definitions := make([]languagesyntax.CaseDefinition, len(rows))
	for index, row := range rows {
		definitions[index] = row.definition
	}
	if !reflect.DeepEqual(definitions, plan.registry.Cases) {
		return profileIssue("REFUTED", "bind-native-cases", "PROFILE_NATIVE_CORPUS_DIVERGED")
	}
	return nil
}

func (plan *ProfileMigrationPlan) replayProfileSources() error {
	names, err := languagesyntax.CaseSourceMembers(plan.registry, plan.request.Case)
	if err != nil {
		return err
	}
	pinned := sourcebundle.Overlay{Files: plan.inputs}
	for index, item := range []struct {
		definition languagesyntax.CaseDefinition
		files      fs.FS
	}{
		{plan.request.Case, pinned},
		{profileAfterCase(plan.request), sourcebundle.Overlay{FS: pinned, Files: map[string][]byte{plan.request.Case.Path: plan.replacement}}},
	} {
		r := languagesyntax.ReplayCaseSources(item.files, item.definition, names)
		if r.ObservedDecision != languagesyntax.DecisionPass || !r.ASTReplayed || !r.ByteReplayed ||
			!r.SemanticReplayed || !r.GetPut || !r.PutGet {
			return fmt.Errorf("PROFILE_SOURCE_CONFORMANCE_FAILED: %s: %v", item.definition.EntityFieldsProfile, r.Diagnostics)
		}
		plan.sourceChecks[index] = r
	}
	return nil
}
