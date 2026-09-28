package provenance

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestReadOnlyMatchesCommittedRead(t *testing.T) {
	store := newReadOnlyWitnessStore(t, false)
	want, err := store.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	before := readOnlyWitnessDisk(t, store.Path())
	got, err := store.ReadOnly(ReadOptions{})
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("read-only observation changed committed evidence: got=%+v err=%v", got, err)
	}
	if !reflect.DeepEqual(before, readOnlyWitnessDisk(t, store.Path())) {
		t.Fatal("read-only observation changed files or metadata")
	}
}

func TestReadOnlyMissingDoesNotCreate(t *testing.T) {
	directory := t.TempDir()
	store := New(filepath.Join(directory, "ledger.jsonl"))
	got, err := store.ReadOnly(ReadOptions{})
	if err != nil || len(got.Records) != 0 {
		t.Fatalf("missing empty store: %+v %v", got, err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("read-only observation created files: %v %v", entries, err)
	}
}

func TestReadOnlyRetainsReadOptionFailures(t *testing.T) {
	store := newReadOnlyWitnessStore(t, true)
	record := BillingFixture()[0]
	cases := []struct {
		name    string
		options ReadOptions
	}{
		{"stale-source", ReadOptions{ExpectedSourceDigest: strings.Repeat("f", 64)}},
		{"expired", ReadOptions{RequireFresh: true, Now: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)}},
		{"candidate-not-verified", ReadOptions{RequiredVerified: []VerifiedClaim{
			{SemanticID: record.SemanticID, SemanticDigest: record.SemanticDigest, GraphDigest: record.GraphDigest},
		}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := readOnlyWitnessDisk(t, store.Path())
			got, err := store.ReadOnly(tc.options)
			if err == nil || errors.Is(err, ErrReadOnlyRecoveryRequired) || got.Digest != "" || len(got.Records) != 0 {
				t.Fatalf("read option failure was hidden or misclassified: %+v %v", got, err)
			}
			if !reflect.DeepEqual(before, readOnlyWitnessDisk(t, store.Path())) {
				t.Fatal("rejected read-only observation changed storage")
			}
		})
	}
}

func TestReadOnlyRejectsMissingStoreIdentity(t *testing.T) {
	for _, store := range []*Store{nil, New("")} {
		if _, err := store.ReadOnly(ReadOptions{}); err == nil {
			t.Fatal("missing store identity was accepted")
		}
	}
}

func newReadOnlyWitnessStore(t *testing.T, candidate bool) *Store {
	t.Helper()
	store := New(filepath.Join(t.TempDir(), "ledger.jsonl"))
	records := BillingFixture()
	if candidate {
		records[0].Status = StatusCandidate
		records[0].Attributes["status"] = string(StatusCandidate)
	}
	if err := store.Append(records...); err != nil {
		t.Fatal(err)
	}
	return store
}

type readOnlyWitnessFile struct {
	Data     string
	Mode     os.FileMode
	Modified time.Time
}

func readOnlyWitnessDisk(t *testing.T, path string) map[string]readOnlyWitnessFile {
	t.Helper()
	directory := filepath.Dir(path)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]readOnlyWitnessFile, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		result[entry.Name()] = readOnlyWitnessFile{Data: string(data), Mode: info.Mode(), Modified: info.ModTime()}
	}
	return result
}
