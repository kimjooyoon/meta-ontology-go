package publictestreuse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
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

// UnmarshalJSON checks the wire cardinality before copying into the fixed array.
// encoding/json otherwise discards excess JSON array elements for Go arrays.
func (inputs *ExecutionPackageInputs) UnmarshalJSON(data []byte) error {
	var wire struct {
		Profile string                  `json:"profile"`
		Members []ExecutionPackageInput `json:"members"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	if len(wire.Members) != len(inputs.Members) {
		return errors.New("generated test package input profile needs exactly three members")
	}
	observed := ExecutionPackageInputs{Profile: wire.Profile}
	copy(observed.Members[:], wire.Members)
	if err := observed.Validate(); err != nil {
		return err
	}
	*inputs = observed
	return nil
}

// ExecutionPackageNames returns the fixed profile in relative-path order.
func ExecutionPackageNames() [3]string {
	return toolchaincli.GeneratedTestPackageNames()
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
