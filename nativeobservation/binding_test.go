package nativeobservation_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/nativeobservation"
)

type wireFixture struct {
	SerializerSource string           `json:"baseline_serializer_source"`
	ReceiptProducer  string           `json:"receipt_producer_source"`
	Bindings         []bindingFixture `json:"bindings"`
}

type bindingFixture struct {
	Name             string `json:"name"`
	ToolchainJSON    string `json:"toolchain_json"`
	ArtifactJSON     string `json:"artifact_json"`
	ExecutableSHA256 string `json:"executable_sha256"`
	ToolchainDigest  string `json:"toolchain_digest"`
	BuildDigest      string `json:"build_digest"`
}

func originalWire(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/original-wire.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestProcessWireMatchesFrozenPreMoveProducer(t *testing.T) {
	var fixture struct {
		Processes []struct {
			Name          string `json:"name"`
			CanonicalJSON string `json:"canonical_json"`
		} `json:"processes"`
	}
	if err := json.Unmarshal(originalWire(t), &fixture); err != nil || len(fixture.Processes) != 7 {
		t.Fatalf("original producer fixture unavailable: %v", err)
	}
	for _, row := range fixture.Processes {
		t.Run(row.Name, func(t *testing.T) {
			var value nativeobservation.ProcessObservation
			if err := json.Unmarshal([]byte(row.CanonicalJSON), &value); err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(value)
			if err != nil || string(got) != row.CanonicalJSON {
				t.Fatalf("field order, omission, bytes, null or int64 changed: %v\n%s", err, got)
			}
		})
	}
}

func TestBindingsMatchFrozenOriginalAndReusedRequests(t *testing.T) {
	var fixture wireFixture
	if err := json.Unmarshal(originalWire(t), &fixture); err != nil || len(fixture.Bindings) != 2 ||
		fixture.SerializerSource != "023cc75ec5869bf3aa6118ae2b5c73f298e797cb" ||
		fixture.ReceiptProducer != "276734167edc7f5fef18cbb79004ea9592553317" {
		t.Fatalf("original binding fixture unavailable: %v", err)
	}
	for _, row := range fixture.Bindings {
		t.Run(row.Name, func(t *testing.T) {
			var reference nativeobservation.ToolchainObservation
			var artifact nativeobservation.ArtifactObservation
			if err := json.Unmarshal([]byte(row.ToolchainJSON), &reference); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(row.ArtifactJSON), &artifact); err != nil {
				t.Fatal(err)
			}
			checkCanonicalJSON(t, reference, row.ToolchainJSON)
			checkCanonicalJSON(t, artifact, row.ArtifactJSON)
			if got := nativeobservation.SourceCheckBindingDigest(reference); got != row.ToolchainDigest {
				t.Fatalf("original source-check bytes changed: %s", got)
			}
			if got := nativeobservation.SourceBuildBindingDigest(artifact.KeySHA256, row.ExecutableSHA256,
				artifact.SourceBuild); got != row.BuildDigest {
				t.Fatalf("original source-build bytes changed: %s", got)
			}
			checkBindingInputs(t, reference, artifact, row)
		})
	}
}

func checkCanonicalJSON(t *testing.T, value any, expected string) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil || string(raw) != expected {
		t.Fatalf("original wrapper wire representation changed: %v\n%s", err, raw)
	}
}

func checkBindingInputs(t *testing.T, reference nativeobservation.ToolchainObservation,
	artifact nativeobservation.ArtifactObservation, row bindingFixture) {
	t.Helper()
	reference.Reused = !reference.Reused
	if nativeobservation.SourceCheckBindingDigest(reference) != row.ToolchainDigest {
		t.Fatal("current reuse flag became original check identity")
	}
	reference.SourceOutput = []byte("different original bytes\n")
	if nativeobservation.SourceCheckBindingDigest(reference) == row.ToolchainDigest {
		t.Fatal("different original version output retained binding")
	}
	artifact.SourceBuild.WallNS++
	if nativeobservation.SourceBuildBindingDigest(artifact.KeySHA256, row.ExecutableSHA256,
		artifact.SourceBuild) == row.BuildDigest {
		t.Fatal("different original build observation retained binding")
	}
}
