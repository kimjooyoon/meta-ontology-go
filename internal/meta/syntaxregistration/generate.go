package syntaxregistration

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
)

// Generate emits all nine members or no candidate. Matching replay bytes prove
// deterministic generation only; existing native conformance still must run.
func (plan Plan) Generate(repository fs.FS) (Candidate, error) {
	if plan.digest == "" || len(plan.binding.activities) != 10 || plan.inputs == nil {
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
	paths := memberPaths(version)
	generators := []func() ([]byte, error){
		func() ([]byte, error) { return generateCorpus(plan.inputs[paths[0]], plan.request) },
		func() ([]byte, error) { return generateRegistry(plan.inputs[paths[1]], plan.request) },
		func() ([]byte, error) { return generateModel(plan.inputs[paths[2]], plan.inputs[corpusPath]) },
		func() ([]byte, error) { return generateSyntaxTests(plan.inputs[paths[3]], plan.inputs[corpusPath]) },
		func() ([]byte, error) { return generateAdmission(plan.inputs[paths[4]], version) },
		func() ([]byte, error) { return generateSelection(plan.inputs[paths[5]], version, capability) },
		func() ([]byte, error) { return generateDigest(plan.inputs[paths[6]], previous, version, next) },
		func() ([]byte, error) { return append([]byte(nil), next...), nil },
		func() ([]byte, error) { return generateMigrationTests(plan.inputs[paths[8]], version, capability) },
	}
	candidate := Candidate{Operation: Operation, ActivityID: plan.binding.activities[0],
		ContractDigest: plan.binding.source, SemanticDigest: plan.binding.semantic,
		InputDigest: plan.digest, RequestDigest: digestValue(plan.request), Toolchain: plan.request.Toolchain,
		State: "PROPOSAL_ONLY", Admission: "UNASSESSED", Required: RequiredMembers, Members: []Member{}}
	for index, generate := range generators {
		raw, err := generate()
		if err != nil {
			if _, ok := errors.AsType[*Failure](err); ok {
				return Candidate{}, err
			}
			return Candidate{}, fmt.Errorf("%w: %v",
				failure("UNKNOWN", "generate:"+paths[index], "REGISTRATION_NATIVE_SHAPE_UNSUPPORTED",
					"UNBOUNDED", "extend-registration-backend"), err)
		}
		before := "ABSENT"
		if old, exists := plan.inputs[paths[index]]; exists {
			before = digest(old)
		}
		candidate.Members = append(candidate.Members, Member{
			paths[index], plan.binding.activities[index+1], before, digest(raw), raw})
	}
	candidate.Emitted = len(candidate.Members)
	return candidate, nil
}

// ValidateCandidate checks the complete candidate against the compiled request,
// current inputs and native generator. It does not authorize project writes.
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
