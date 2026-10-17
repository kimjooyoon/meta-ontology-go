package publictestreuse

import (
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const ExecutionPackageProfile = "gooo/generated-test-package-inputs/v1"

// ExecutionPackageInputs describes the three files in the existing generated
// test package. Tool, environment and transitive input identity are separate.
type ExecutionPackageInputs struct {
	Profile string `json:"profile"`

	Members [3]ExecutionPackageInput `json:"members"`
}

// ExecutionPackageInput retains a file's role and raw-byte identity.
type ExecutionPackageInput struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	Bytes  int64  `json:"bytes"`
	Digest string `json:"digest"`
}

// ExecutionPackageNames returns the fixed profile in relative-path order.
func ExecutionPackageNames() [3]string {
	return [3]string{"generated_project_test.go", "go.mod", "semantic.gooo.go"}
}

func executionPackageRoles() [3]string {
	return [3]string{"test-contract", "module", "generated-program"}
}

// ExecutionPackageInputsFromBytes records caller-supplied bytes. The filesystem
// reader must establish membership and file kinds before using this profile.
func ExecutionPackageInputsFromBytes(module, program, contract []byte) ExecutionPackageInputs {
	result := ExecutionPackageInputs{Profile: ExecutionPackageProfile}
	names, roles := ExecutionPackageNames(), executionPackageRoles()
	sources := [3][]byte{contract, module, program}
	for index, data := range sources {
		result.Members[index] = ExecutionPackageInput{
			Path:   names[index],
			Role:   roles[index],
			Bytes:  int64(len(data)),
			Digest: cache.HashBytes(data).String(),
		}
	}
	return result
}

// Validate checks representation. Matching current bytes requires a fresh
// reader result; this method alone does not establish a filesystem snapshot.
func (inputs ExecutionPackageInputs) Validate() error {
	if inputs.Profile != ExecutionPackageProfile {
		return errors.New("generated test package input profile is unknown")
	}
	names, roles := ExecutionPackageNames(), executionPackageRoles()
	for index, member := range inputs.Members {
		if member.Path != names[index] || member.Role != roles[index] || member.Bytes < 0 ||
			!cache.Digest(member.Digest).Known() {
			return fmt.Errorf("generated test package input member %d is invalid", index)
		}
	}
	return nil
}
