//go:build generated_project
// +build generated_project

package publicdiscoveryexample

import "testing"

func TestGeneratedProjectContract(t *testing.T) {
	if got := Compile(Input{}); got != (Output{}) {
		t.Fatalf("generated Compile returned %#v, want an empty Output", got)
	}
}
