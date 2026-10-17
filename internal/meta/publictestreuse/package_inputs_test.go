package publictestreuse

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestExecutionPackageInputsBindRawBytesAndRoles(t *testing.T) {
	module, program, contract := []byte("module fixture\n"), []byte("package fixture\n"), []byte("package fixture\n// test\n")
	inputs := ExecutionPackageInputsFromBytes(module, program, contract)
	if err := inputs.Validate(); err != nil {
		t.Fatal(err)
	}
	wantPaths := [3]string{"generated_project_test.go", "go.mod", "semantic.gooo.go"}
	wantRoles := [3]string{"test-contract", "module", "generated-program"}
	wantBytes := [3][]byte{contract, module, program}
	for index, member := range inputs.Members {
		sum := sha256.Sum256(wantBytes[index])
		if member.Path != wantPaths[index] || member.Role != wantRoles[index] ||
			member.Bytes != int64(len(wantBytes[index])) || member.Digest != fmt.Sprintf("%x", sum) {
			t.Fatalf("member %d = %+v", index, member)
		}
	}
	changed := ExecutionPackageInputsFromBytes(append(module, '\n'), program, contract)
	if changed == inputs || changed.Members[0] != inputs.Members[0] || changed.Members[2] != inputs.Members[2] {
		t.Fatal("module changes must alter only the module input identity")
	}
	copy := inputs
	copy.Members[0].Role = "changed"
	if inputs.Members[0].Role != "test-contract" {
		t.Fatal("fixed member array must preserve value-copy isolation")
	}
}

func TestExecutionPackageInputsRejectIncompleteProfile(t *testing.T) {
	inputs := ExecutionPackageInputsFromBytes([]byte("module fixture\n"), []byte("package fixture\n"), []byte("package fixture\n"))
	for _, name := range []string{"profile", "missing", "path", "role", "bytes", "digest", "order"} {
		t.Run(name, func(t *testing.T) {
			changed := inputs
			switch name {
			case "profile":
				changed.Profile = ""
			case "missing":
				changed.Members[1] = ExecutionPackageInput{}
			case "path":
				changed.Members[1].Path = "other.mod"
			case "role":
				changed.Members[1].Role = "display"
			case "bytes":
				changed.Members[1].Bytes = -1
			case "digest":
				changed.Members[1].Digest = ""
			case "order":
				changed.Members[0], changed.Members[1] = changed.Members[1], changed.Members[0]
			}
			if changed.Validate() == nil {
				t.Fatal("invalid input profile was accepted")
			}
		})
	}
}
