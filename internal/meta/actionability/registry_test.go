package actionability

import "testing"

func TestPartitionDirectoryExecutorRegistered(t *testing.T) {
	index, err := executorIndex(canonicalExecutors())
	if err != nil {
		t.Fatal(err)
	}
	got, ok := index["partition-directory"]
	if !ok {
		t.Fatal("partition-directory executor is not registered")
	}
	want := Executor{
		Operation: "partition-directory", Activity: "PartitionDirectory",
		ProofChoice: "foundation", Registry: "source-policy",
		Executor: "cmd/directory-partition-witness", Evaluator: "cmd/directory-partition-witness:check",
	}
	if got != want {
		t.Fatalf("partition-directory executor = %#v, want %#v", got, want)
	}
}

func TestSeparateDirectoryKindsExecutorRegistered(t *testing.T) {
	index, err := executorIndex(canonicalExecutors())
	if err != nil {
		t.Fatal(err)
	}
	got, ok := index["separate-directory-kinds"]
	if !ok {
		t.Fatal("separate-directory-kinds executor is not registered")
	}
	want := Executor{
		Operation: "separate-directory-kinds", Activity: "SeparateDirectoryKinds",
		ProofChoice: "foundation", Registry: "source-policy",
		Executor: "cmd/directory-kind-witness", Evaluator: "cmd/directory-kind-witness:check",
	}
	if got != want {
		t.Fatalf("separate-directory-kinds executor = %#v, want %#v", got, want)
	}
}
