package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func TestObservedAdmissionOriginalReferencesFresh1901(t *testing.T) {
	f, baseline := admissionFixture1894(t)
	for _, mode := range []struct {
		id, kind       string
		members, files int
	}{
		{"baseline", "", 11, 4},
		{"document-bytes", "DOCUMENT_BYTES", 11, 4},
		{"missing-observation-inputs", "MISSING_OBSERVATION_INPUTS", 9, 2},
		{"legacy-downgrade", "LEGACY_DOWNGRADE", 11, 3},
	} {
		t.Run(mode.id, func(t *testing.T) {
			v := baseline
			if mode.kind != "" {
				var err error
				v, err = observedAdmissionVariant(baseline, f, countercase{ID: mode.id, Kind: mode.kind})
				if err != nil {
					t.Fatal(err)
				}
			}
			directory, source := t.TempDir(), origin{ArtifactID: 88, RunID: 77, Attempt: 1}
			if err := saveVariant(directory, source, f, v); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(directory, "inputs.json"))
			if err != nil {
				t.Fatal(err)
			}
			var manifest struct {
				Origin origin
				Inputs []inputReference
			}
			if err := json.Unmarshal(raw, &manifest); err != nil || manifest.Origin != source || len(manifest.Inputs) != mode.members {
				t.Fatal("original transport identity or input population changed", err)
			}
			for _, input := range manifest.Inputs {
				changed := input.Member == "operation-observations.json" || input.Member == "observation-policy.gooo" ||
					input.Member == "concept-operation-binding.json" && mode.id != "legacy-downgrade"
				want := "ORIGINAL_PRODUCER_ARCHIVE"
				if changed {
					want = "CHANGED_MEMBER"
					payload, err := os.ReadFile(filepath.Join(directory, input.Member))
					if err != nil || artifact.ContentDigest(payload) != input.Digest {
						t.Fatal("changed payload lost its exact bytes", err)
					}
				}
				if input.State != want {
					t.Fatalf("new observed wrapper claimed original archive authority: %+v", input)
				}
			}
			files, err := os.ReadDir(directory)
			if err != nil || len(files) != mode.files {
				t.Fatal("unchanged original payloads were copied", err)
			}
		})
	}
}
