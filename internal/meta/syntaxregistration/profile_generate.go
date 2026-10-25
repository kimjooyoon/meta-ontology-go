package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"slices"
)

func (plan ProfileMigrationPlan) Generate(repository fs.FS) (ProfileMigrationCandidate, error) {
	inputs, err := readProfileInputs(repository, plan.request)
	if err != nil {
		return ProfileMigrationCandidate{}, err
	}
	if digestValue(inputs) != plan.request.SnapshotDigest {
		return ProfileMigrationCandidate{}, profileIssue("UNKNOWN", "recheck-input", "PROFILE_MIGRATION_INPUT_STALE")
	}
	if err := recheckExecutionIdentity(plan.request.ExecutionIdentity); err != nil {
		return ProfileMigrationCandidate{}, err
	}
	corpus, err := plan.generateProfileCorpus()
	if err != nil {
		return ProfileMigrationCandidate{}, err
	}
	nativePath, native, err := plan.generateProfileNative()
	if err != nil {
		return ProfileMigrationCandidate{}, err
	}
	paths := [3]string{corpusPath, nativePath, plan.request.Case.Path}
	contents := [3][]byte{corpus, native, plan.replacement}
	if paths[0] == paths[1] || paths[0] == paths[2] || paths[1] == paths[2] {
		return ProfileMigrationCandidate{}, fmt.Errorf("PROFILE_OUTPUT_ROLE_COLLISION")
	}
	candidate := ProfileMigrationCandidate{Schema: "gooo/syntax-profile-migration-candidate/v1",
		Operation: "MigrateSyntaxProfile", ActivityID: plan.binding.operation,
		ContractDigest: plan.binding.source, SemanticDigest: plan.binding.semantic,
		InputDigest: plan.request.SnapshotDigest, RequestDigest: digestValue(plan.request),
		BeforeCase: plan.request.Case, AfterCase: profileAfterCase(plan.request),
		SourceChecks:   plan.sourceChecks,
		CasePopulation: len(plan.registry.Cases), RequiredMembers: 3, Admission: "UNASSESSED", State: "PROPOSAL_ONLY",
		ExecutionBinding: ExecutionBinding{ActivityID: plan.binding.pin,
			InputID: profileMigrationDomain + "request", OutputID: profileMigrationDomain + "pinned-input",
			Identity: plan.request.ExecutionIdentity}}
	for index, check := range candidate.SourceChecks {
		candidate.SourceChecks[index].Diagnostics = append([]string{}, check.Diagnostics...)
		candidate.SourceChecks[index].SourceMembers = slices.Clone(check.SourceMembers)
	}
	for index, path := range paths {
		candidate.Members[index] = ProfileMigrationMember{Path: path,
			ActivityID: plan.binding.activities[index], OutputID: plan.binding.outputs[index],
			BeforeDigest: digest(plan.inputs[path]), AfterDigest: digest(contents[index]),
			BeforeContent: bytes.Clone(plan.inputs[path]), Content: bytes.Clone(contents[index])}
	}
	return candidate, nil
}

func (plan ProfileMigrationPlan) ValidateCandidate(repository fs.FS, candidate ProfileMigrationCandidate) error {
	expected, err := plan.Generate(repository)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, candidate) {
		return profileIssue("REFUTED", "validate-candidate", "PROFILE_CANDIDATE_DIVERGED")
	}
	return nil
}

func DecodeProfileMigrationCandidate(raw []byte) (ProfileMigrationCandidate, error) {
	var candidate ProfileMigrationCandidate
	err := decodeStrict(raw, &candidate)
	return candidate, err
}

func (plan ProfileMigrationPlan) generateProfileCorpus() ([]byte, error) {
	raw := plan.inputs[corpusPath]
	var document map[string]json.RawMessage
	if err := decodeStrict(raw, &document); err != nil {
		return nil, err
	}
	array := document["cases"]
	var rows []json.RawMessage
	if err := decodeStrict(array, &rows); err != nil {
		return nil, err
	}
	old := rows[plan.index]
	if bytes.Count(raw, array) != 1 || bytes.Count(array, old) != 1 {
		return nil, fmt.Errorf("PROFILE_CORPUS_CASE_ORIGIN_AMBIGUOUS")
	}
	next, err := json.MarshalIndent(profileAfterCase(plan.request), "    ", "  ")
	if err != nil {
		return nil, err
	}
	updated := bytes.Replace(array, old, next, 1)
	return bytes.Replace(raw, array, updated, 1), nil
}
