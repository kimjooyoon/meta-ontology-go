package syntaxregistration

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"slices"
)

// Generate closes all nine semantic artifact roles or emits no candidate.
// Physical member paths come from exact AST ownership, not filename guesses.
func (plan Plan) Generate(repository fs.FS) (Candidate, error) {
	if plan.digest == "" || len(plan.binding.activities) != 10 || len(plan.binding.outputs) != 10 || plan.inputs == nil {
		return Candidate{}, failure("REFUTED", "validate-plan", "REGISTRATION_PLAN_NOT_COMPILED", "", "compile-source-bound-plan")
	}
	current, err := readInputs(repository, plan.request)
	if err != nil {
		return Candidate{}, err
	}
	if digestValue(current) != plan.digest {
		return Candidate{}, failure("UNKNOWN", "recheck-snapshot", "REGISTRATION_INPUT_STALE", "STALE", "recompile-exact-input")
	}
	_, _, capability, err := corpusTotals(plan.inputs[corpusPath])
	if err != nil {
		return Candidate{}, err
	}
	version := plan.request.BaseVersion + 1
	previous := plan.inputs[denominatorPath(version-1)]
	next, err := generateDenominator(previous, version-1, capability)
	if err != nil {
		return Candidate{}, failure("REFUTED", "bind-denominator", "REGISTRATION_BASELINE_DENOMINATOR_MISMATCH", "", "restore-bound-baseline")
	}
	syntax, err := parseSourceUnits(plan.inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		return Candidate{}, err
	}
	conformance, err := parseSourceUnits(plan.inputs, syntaxRoot+"conformance/", "languagesyntax_test", true)
	if err != nil {
		return Candidate{}, err
	}
	closure, err := parseSourceUnits(plan.inputs, closureRoot, "verticalsliceclosureshadow", true)
	if err != nil {
		return Candidate{}, err
	}
	generated := map[string][]byte{denominatorPath(version): next}
	generated[corpusPath], err = generateCorpus(plan.inputs[corpusPath], plan.request)
	if err != nil {
		return Candidate{}, err
	}
	rolePaths := make([]map[string]bool, RequiredArtifacts)
	for index := range rolePaths {
		rolePaths[index] = map[string]bool{}
	}
	rolePaths[0][corpusPath] = true
	rolePaths[7][denominatorPath(version)] = true
	steps := []struct {
		role   int
		source *goSource
		run    func() error
	}{
		{1, syntax, func() error { return generateRegistry(syntax, plan.request) }},
		{2, syntax, func() error { return generateModel(syntax, plan.inputs[corpusPath]) }},
		{3, conformance, func() error { return generateSyntaxTests(conformance, plan.inputs[corpusPath]) }},
		{4, closure, func() error { return generateAdmission(closure, version) }},
		{5, closure, func() error { return generateSelection(closure, version, capability) }},
		{6, closure, func() error { return generateDigest(closure, previous, version, next) }},
		{8, closure, func() error { return generateMigrationTests(closure, version, capability) }},
	}
	for _, step := range steps {
		step.source.activity = plan.binding.activities[step.role+1]
		before := len(step.source.edits)
		if err := step.run(); err != nil {
			if _, ok := errors.AsType[*Failure](err); ok {
				return Candidate{}, err
			}
			return Candidate{}, fmt.Errorf("%w: %v", failure("UNKNOWN",
				"generate:"+plan.binding.outputs[step.role+1], "REGISTRATION_NATIVE_SHAPE_UNSUPPORTED",
				"UNBOUNDED", "extend-registration-backend"), err)
		}
		for _, edit := range step.source.edits[before:] {
			rolePaths[step.role][edit.path] = true
		}
	}
	for _, source := range []*goSource{syntax, conformance, closure} {
		changes, err := source.finish()
		if err != nil {
			return Candidate{}, err
		}
		for path, raw := range changes {
			if _, duplicate := generated[path]; duplicate {
				return Candidate{}, fmt.Errorf("registration source ownership overlaps: %s", path)
			}
			generated[path] = raw
		}
	}
	candidate := Candidate{Operation: Operation, ActivityID: plan.binding.activities[0],
		ContractDigest: plan.binding.source, SemanticDigest: plan.binding.semantic,
		InputDigest: plan.digest, RequestDigest: digestValue(plan.request), Toolchain: plan.request.Toolchain,
		State: "PROPOSAL_ONLY", Admission: "UNASSESSED", RequiredArtifacts: RequiredArtifacts,
		Required: len(generated), Members: []Member{}, Artifacts: []Artifact{}}
	for index, paths := range rolePaths {
		if len(paths) == 0 {
			return Candidate{}, failure("UNKNOWN", "bind-artifact:"+plan.binding.outputs[index+1],
				"REGISTRATION_ARTIFACT_UNRESOLVED", "DIRECT_MISSING", "restore-role-source-binding")
		}
		candidate.Artifacts = append(candidate.Artifacts, Artifact{
			plan.binding.activities[index+1], plan.binding.outputs[index+1], sortedPaths(paths)})
	}
	for _, path := range sortedPaths(generated) {
		raw := generated[path]
		member := Member{Path: path, BeforeDigest: "ABSENT", AfterDigest: digest(raw), Content: raw}
		if old, exists := plan.inputs[path]; exists {
			member.BeforeDigest = digest(old)
		}
		for _, artifact := range candidate.Artifacts {
			if slices.Contains(artifact.Paths, path) {
				member.ActivityIDs = append(member.ActivityIDs, artifact.ActivityID)
			}
		}
		if len(member.ActivityIDs) == 1 {
			member.ActivityID = member.ActivityIDs[0]
		}
		candidate.Members = append(candidate.Members, member)
	}
	candidate.Emitted = len(candidate.Members)
	return candidate, nil
}

// ValidateCandidate checks the entire write set and all nine lowered bindings.
// It grants neither semantic acceptance nor application authority.
func (plan Plan) ValidateCandidate(repository fs.FS, candidate Candidate) error {
	expected, err := plan.Generate(repository)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(candidate, expected) {
		return failure("REFUTED", "validate-candidate", "REGISTRATION_CANDIDATE_MISMATCH", "", "report-counterexample")
	}
	return nil
}
