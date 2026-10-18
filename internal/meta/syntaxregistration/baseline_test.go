package syntaxregistration

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"testing/fstest"
)

func TestCurrentBaselinePinsTheComposedGoooCase(t *testing.T) {
	data, original := fixture(t)
	version, err := ObserveCurrentBaseline(data)
	if err != nil || version != original.BaseVersion {
		t.Fatalf("current baseline differs from complete history: %d %v", version, err)
	}
	intent := CaseIntent{ID: original.Case.ID, Path: original.Case.Path,
		ProofChoice: original.Case.ProofChoice, Profile: "text-fields"}
	request, value, err := ComposeCurrentDraft(context.Background(), data, intent)
	if err != nil || request.BaseVersion != version || request.CaseIntent == nil ||
		request.CaseIntent.BaseVersion != version || request.CaseSourceDigest != value.SourceSHA256 {
		t.Fatalf("observed baseline was not pinned to Gooo intent: %+v %v", request, err)
	}
	intent.BaseVersion = version
	if _, _, err := ComposeCurrentDraft(context.Background(), data, intent); err != nil {
		t.Fatalf("explicit current version was refused: %v", err)
	}
	intent.BaseVersion--
	if _, _, err := ComposeCurrentDraft(context.Background(), data, intent); failureReason(err) != "REGISTRATION_BASELINE_VERSION_STALE" {
		t.Fatalf("older explicit version was not stale: %v", err)
	}
}

func TestCurrentBaselineRefusesBrokenHistory(t *testing.T) {
	for _, row := range []struct {
		name   string
		change func(t *testing.T, data fstest.MapFS, version int)
		want   string
	}{
		{"missing-middle", func(t *testing.T, data fstest.MapFS, version int) {
			delete(data, denominatorPath(version-1))
		}, "REGISTRATION_HISTORY_NOT_CONTIGUOUS"},
		{"noncanonical-duplicate", func(t *testing.T, data fstest.MapFS, version int) {
			data[closureRoot+"evidence/denominator-v0"+strconv.Itoa(version)+".json"] = data[denominatorPath(version)]
		}, "REGISTRATION_HISTORY_NOT_CONTIGUOUS"},
		{"wrong-header", func(t *testing.T, data fstest.MapFS, version int) {
			changeDenominator(t, data, version, func(value *denominator) { value.Version-- })
		}, "REGISTRATION_HISTORY_CONTRADICTION"},
		{"lowered-history-target", func(t *testing.T, data fstest.MapFS, version int) {
			changeDenominator(t, data, version-1, func(value *denominator) { value.Boundaries[0].Target = 1 })
		}, "REGISTRATION_HISTORY_CONTRADICTION"},
		{"lowered-link-obligation", func(t *testing.T, data fstest.MapFS, version int) {
			changeDenominator(t, data, version, func(value *denominator) { value.Boundaries[0].LinkTarget = 0 })
		}, "REGISTRATION_HISTORY_CONTRADICTION"},
		{"current-corpus-mismatch", func(t *testing.T, data fstest.MapFS, version int) {
			changeDenominator(t, data, version, func(value *denominator) { value.Boundaries[0].Target++ })
		}, "REGISTRATION_BASELINE_DENOMINATOR_MISMATCH"},
		{"missing-corpus", func(t *testing.T, data fstest.MapFS, version int) {
			delete(data, corpusPath)
		}, "REGISTRATION_CORPUS_UNREADABLE"},
	} {
		t.Run(row.name, func(t *testing.T) {
			data, original := fixture(t)
			row.change(t, data, original.BaseVersion)
			if _, err := ObserveCurrentBaseline(data); failureReason(err) != row.want {
				t.Fatalf("broken baseline was not refused as %s: %v", row.want, err)
			}
		})
	}
}

func changeDenominator(t *testing.T, data fstest.MapFS, version int, change func(*denominator)) {
	t.Helper()
	path := denominatorPath(version)
	var value denominator
	if err := json.Unmarshal(data[path].Data, &value); err != nil {
		t.Fatal(err)
	}
	change(&value)
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	data[path] = &fstest.MapFile{Data: raw}
}

func failureReason(err error) string {
	var observed *Failure
	if errors.As(err, &observed) {
		return observed.Reason
	}
	return ""
}
