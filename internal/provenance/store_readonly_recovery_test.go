package provenance

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestReadOnlyDoesNotRecover(t *testing.T) {
	cases := []struct {
		name   string
		reason string
		count  int
	}{
		{"prepared", ReadOnlyPreparedTransaction, 1},
		{"committed-divergence", ReadOnlyMaterializationMismatch, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := newReadOnlyWitnessStore(t, false)
			original, err := store.Read(ReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "prepared" {
				readOnlyWitnessPrepared(t, store, original, false)
			} else if err := os.WriteFile(store.Path(), []byte("partial"), 0o644); err != nil {
				t.Fatal(err)
			}
			before := readOnlyWitnessDisk(t, store.Path())
			got, err := store.ReadOnly(ReadOptions{})
			var recovery *ReadOnlyRecoveryError
			if !errors.Is(err, ErrReadOnlyRecoveryRequired) || !errors.As(err, &recovery) {
				t.Fatalf("missing explicit recovery cause: %v", err)
			}
			if recovery.Path != store.Path() || recovery.Reason != tc.reason || got.Digest != "" || len(got.Records) != 0 {
				t.Fatalf("recovery boundary changed: %+v %+v", recovery, got)
			}
			if !reflect.DeepEqual(before, readOnlyWitnessDisk(t, store.Path())) {
				t.Fatal("observation repaired or rolled back the original storage")
			}
			repaired, err := store.Read(ReadOptions{})
			if err != nil || !reflect.DeepEqual(repaired.Records, original.Records[:tc.count]) {
				t.Fatalf("existing explicit recovery changed: %+v %v", repaired, err)
			}
		})
	}
}

func TestReadOnlyDoesNotDowngradeContradictions(t *testing.T) {
	for _, name := range []string{"committed-digest-conflict", "prepared-not-append"} {
		t.Run(name, func(t *testing.T) {
			store := newReadOnlyWitnessStore(t, false)
			original, err := store.Read(ReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if name == "prepared-not-append" {
				readOnlyWitnessPrepared(t, store, original, true)
			} else {
				manifest, err := readManifest(store.Path())
				if err != nil {
					t.Fatal(err)
				}
				manifest.Digest = strings.Repeat("0", 64)
				readOnlyWitnessWriteManifest(t, store, manifest)
			}
			before := readOnlyWitnessDisk(t, store.Path())
			got, err := store.ReadOnly(ReadOptions{})
			if err == nil || errors.Is(err, ErrReadOnlyRecoveryRequired) || got.Digest != "" || len(got.Records) != 0 {
				t.Fatalf("contradictory metadata was accepted or softened: %+v %v", got, err)
			}
			if !reflect.DeepEqual(before, readOnlyWitnessDisk(t, store.Path())) {
				t.Fatal("contradictory evidence was changed during observation")
			}
		})
	}
}

func readOnlyWitnessPrepared(t *testing.T, store *Store, original Snapshot, backwards bool) {
	t.Helper()
	full, err := canonicalJSONL(original.Records)
	if err != nil {
		t.Fatal(err)
	}
	prefix, err := canonicalJSONL(original.Records[:1])
	if err != nil {
		t.Fatal(err)
	}
	manifest := preparedManifestFor(prefix, original.Records[:1], full, original.Records)
	if backwards {
		manifest = preparedManifestFor(full, original.Records, prefix, original.Records[:1])
	}
	readOnlyWitnessWriteManifest(t, store, manifest)
}

func readOnlyWitnessWriteManifest(t *testing.T, store *Store, manifest ledgerManifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath(store.Path()), data, 0o644); err != nil {
		t.Fatal(err)
	}
}
