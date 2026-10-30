package nativeobservation_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/nativeobservation"
)

func TestSourceOutputKeepsNilEmptyAndUntrimmedBytes(t *testing.T) {
	var fixture struct {
		Outputs []struct {
			Name           string `json:"name"`
			SourceOutput   []byte `json:"source_output"`
			CanonicalTuple string `json:"canonical_tuple"`
			Digest         string `json:"digest"`
		} `json:"byte_outputs"`
	}
	if err := json.Unmarshal(originalWire(t), &fixture); err != nil || len(fixture.Outputs) != 3 {
		t.Fatalf("literal byte-domain oracle unavailable: %v", err)
	}
	for _, row := range fixture.Outputs {
		t.Run(row.Name, func(t *testing.T) {
			want := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(row.CanonicalTuple)))
			if want != row.Digest {
				t.Fatal("literal original tuple bytes no longer match frozen digest")
			}
			value := nativeobservation.ToolchainObservation{SourceOutput: row.SourceOutput}
			if nativeobservation.SourceCheckBindingDigest(value) != row.Digest {
				t.Fatal("nil/empty distinction or exact source bytes changed")
			}
		})
	}
}
