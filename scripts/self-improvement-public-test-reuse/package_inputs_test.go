package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecutionPackageInputsMapPublishedCopies(t *testing.T) {
	root, module, program, contract := writePackageInputFixture(t, false)
	observed, err := readExecutionPackageInputs(root, program, contract)
	if err != nil {
		t.Fatal(err)
	}
	published, _, _, _ := writePackageInputFixture(t, true)
	if err := os.WriteFile(filepath.Join(published, "explanation.md"), []byte("display only\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mapped, err := readPublishedPackageInputs(published, program, contract)
	if err != nil || mapped != observed {
		t.Fatalf("published mapping differs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), append(module, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := readExecutionPackageInputs(root, program, contract)
	if err != nil || changed == observed || changed.Members[0] != observed.Members[0] || changed.Members[2] != observed.Members[2] {
		t.Fatalf("module bytes were not bound independently: %v", err)
	}
}

func TestExecutionPackageInputsRejectUnobservedMembers(t *testing.T) {
	for _, name := range []string{"extra", "missing", "renamed", "directory", "program", "contract"} {
		t.Run(name, func(t *testing.T) {
			root, _, program, contract := writePackageInputFixture(t, false)
			var err error
			switch name {
			case "extra":
				err = os.WriteFile(filepath.Join(root, "extra.go"), []byte("package fixture\n"), 0o600)
			case "missing":
				err = os.Remove(filepath.Join(root, "go.mod"))
			case "renamed":
				err = os.Rename(filepath.Join(root, "go.mod"), filepath.Join(root, "other.mod"))
			case "directory":
				if err = os.Remove(filepath.Join(root, "go.mod")); err == nil {
					err = os.Mkdir(filepath.Join(root, "go.mod"), 0o700)
				}
			case "program":
				err = os.WriteFile(filepath.Join(root, "semantic.gooo.go"), []byte("different"), 0o600)
			case "contract":
				err = os.WriteFile(filepath.Join(root, "generated_project_test.go"), []byte("different"), 0o600)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := readExecutionPackageInputs(root, program, contract); err == nil {
				t.Fatal("unobserved execution package inputs were accepted")
			}
		})
	}
}

func writePackageInputFixture(t *testing.T, published bool) (string, []byte, []byte, []byte) {
	t.Helper()
	root := t.TempDir()
	module, program, contract := []byte("module fixture\n"), []byte("package fixture\n"), []byte("package fixture\n// test\n")
	names := [3]string{"generated_project_test.go", "go.mod", "semantic.gooo.go"}
	if published {
		names = [3]string{"test-reuse-test-contract.go", "go.mod", "test-reuse-baseline-program.go"}
	}
	for index, data := range [3][]byte{contract, module, program} {
		if err := os.WriteFile(filepath.Join(root, names[index]), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root, module, program, contract
}
