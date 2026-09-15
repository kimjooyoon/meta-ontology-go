package verify

import (
	"path/filepath"
	"testing"
)

func TestGovernanceMatrixRejectsShrunkenKernel(t *testing.T) {
	matrix, err := ReadGovernanceMatrix(filepath.Join("..", "..", ".github", "ci-governance.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range mandatoryKernelPaths() {
		shrunk := matrix
		shrunk.ProtectedKernel = nil
		for _, path := range matrix.ProtectedKernel {
			if path != required {
				shrunk.ProtectedKernel = append(shrunk.ProtectedKernel, path)
			}
		}
		if err := ValidateGovernanceMatrix(shrunk); err == nil {
			t.Fatalf("kernel path %q could be removed without rejection", required)
		}
	}
}
