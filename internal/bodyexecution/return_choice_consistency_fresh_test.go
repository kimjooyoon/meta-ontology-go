package bodyexecution

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadReturnChoiceRuntime22217(t *testing.T, name string) CompositionRuntime {
	t.Helper()
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	file, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var runtime CompositionRuntime
	if err := json.NewDecoder(z).Decode(&runtime); err != nil {
		t.Fatal(err)
	}
	return runtime
}

// Refresh every affected self-digest: rejection must follow structural
// contradictions, rather than merely noticing an old checksum.
func refreshReturnChoiceDigests22217(r *CompositionRuntime) {
	r.FieldObservation.MetadataSHA256 = compositionDigest([]any{r.FieldProjectionSHA256, r.FieldBindings, r.FieldSites, r.FieldUseSites})
	r.FieldSources.SHA256 = fieldSourcesDigest(*r.FieldSources)
	r.FieldFileSources.FieldSourcesSHA256 = r.FieldSources.SHA256
	r.FieldFileSources.SHA256 = fieldFileSourcesDigest(*r.FieldFileSources)
}

func TestReturnChoiceRecordedConsistency22217(t *testing.T) {
	for _, fixture := range []string{"choice22133-runtime.json.gz", "return21917-runtime.json.gz"} {
		r := loadReturnChoiceRuntime22217(t, fixture)
		if err := validateRecordedNativeOrigins(r); err != nil {
			t.Fatal("unchanged original record", fixture, err)
		}
	}
	choiceCases := []struct {
		name string
		edit func(*CompositionFieldChoiceSource)
		want string
	}{
		{"schema", func(c *CompositionFieldChoiceSource) { c.Schema = "other" }, "source identity"},
		{"missing-choice", func(c *CompositionFieldChoiceSource) { c.ChoiceID = "" }, "source identity"},
		{"option", func(c *CompositionFieldChoiceSource) { c.OptionID = "third" }, "source identity"},
		{"kind", func(c *CompositionFieldChoiceSource) { c.Kind = "field_update" }, "typed site"},
		{"record", func(c *CompositionFieldChoiceSource) { c.RecordID += ":other" }, "typed site"},
		{"field", func(c *CompositionFieldChoiceSource) { c.FieldID += ":other" }, "typed site"},
		{"selected-body", func(c *CompositionFieldChoiceSource) { c.SelectedBodySHA += ":other" }, "selected expression"},
		{"selected-start", func(c *CompositionFieldChoiceSource) { c.SelectedDecoded.Start++ }, "selected expression"},
		{"declaration-view", func(c *CompositionFieldChoiceSource) { c.Declaration.View = "other" }, "view differs"},
		{"declaration-range", func(c *CompositionFieldChoiceSource) { c.Declaration.End = c.Declaration.Start }, "incomplete"},
		{"value-range", func(c *CompositionFieldChoiceSource) { c.Value.Start = -1 }, "incomplete"},
		{"empty-expression", func(c *CompositionFieldChoiceSource) { c.Value.Expression = "" }, "expression absent"},
		{"original-source", func(c *CompositionFieldChoiceSource) { c.Value.SourceSHA256 += ":other" }, "origins differ"},
		{"original-package", func(c *CompositionFieldChoiceSource) { c.Value.PackagePath += "/other" }, "origins differ"},
		{"original-file", func(c *CompositionFieldChoiceSource) { c.Value.Filename += ".other" }, "origins differ"},
		{"original-activity", func(c *CompositionFieldChoiceSource) { c.Value.ActivityID += ":other" }, "origins differ"},
	}
	for _, tc := range choiceCases {
		t.Run("choice/"+tc.name, func(t *testing.T) {
			r := loadReturnChoiceRuntime22217(t, "choice22133-runtime.json.gz")
			tc.edit(r.FieldFileSources.Bindings[0].Choice)
			refreshReturnChoiceDigests22217(&r)
			if err := validateRecordedNativeOrigins(r); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatal("recomputed choice contradiction", err, tc.want)
			}
		})
	}
	returnCases := []struct {
		name string
		edit func(*CompositionFieldBinding)
		want string
	}{
		{"kind", func(b *CompositionFieldBinding) { b.ExpressionKind = "expression" }, "return-expression"},
		{"ordinal", func(b *CompositionFieldBinding) { b.ExpressionOrdinal = -1 }, "return-expression"},
		{"name", func(b *CompositionFieldBinding) { b.Name = "invented" }, "return-expression"},
		{"input", func(b *CompositionFieldBinding) { b.InputKind = "root_input" }, "return-expression"},
		{"missing-declaration", func(b *CompositionFieldBinding) { b.Declaration = nil }, "return-expression"},
		{"selected-source", func(b *CompositionFieldBinding) { b.Declaration.SourceSHA256 += ":other" }, "declaration source"},
		{"view", func(b *CompositionFieldBinding) { b.Declaration.View = "other" }, "declaration source"},
		{"range", func(b *CompositionFieldBinding) { b.Declaration.End = b.Declaration.Start }, "declaration is incomplete"},
		{"decoded-range", func(b *CompositionFieldBinding) { b.Declaration.DecodedEnd = b.Declaration.DecodedStart }, "declaration is incomplete"},
		{"empty-expression", func(b *CompositionFieldBinding) { b.Declaration.Expression = "" }, "declaration is incomplete"},
		{"mapping", func(b *CompositionFieldBinding) { b.Declaration.Status = "other" }, "mapping differs"},
		{"unmapped-range", func(b *CompositionFieldBinding) { b.Declaration.Status = "UNMAPPED"; b.Declaration.Reason = "absent" }, "claims a range"},
	}
	for _, tc := range returnCases {
		t.Run("return/"+tc.name, func(t *testing.T) {
			r := loadReturnChoiceRuntime22217(t, "return21917-runtime.json.gz")
			for i := range r.FieldBindings {
				if r.FieldBindings[i].ExpressionKind == "return" {
					tc.edit(&r.FieldBindings[i])
					break
				}
			}
			refreshReturnChoiceDigests22217(&r)
			if err := validateRecordedNativeOrigins(r); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatal("recomputed return contradiction", err, tc.want)
			}
		})
	}
	for _, wrongField := range []bool{false, true} {
		r := loadReturnChoiceRuntime22217(t, "return21917-runtime.json.gz")
		for i := range r.Traces[0].FieldWrites {
			w := &r.Traces[0].FieldWrites[i]
			if w.ParentAssignment >= 0 {
				if wrongField {
					w.ParentAssignment = (w.ParentAssignment + 1) % 2
				} else {
					w.ParentAssignment = i
				}
				break
			}
		}
		refreshReturnChoiceDigests22217(&r)
		if err := validateRecordedNativeOrigins(r); err == nil || !strings.Contains(err.Error(), "parent identity or completion") {
			t.Fatal("contradictory helper return parent", wrongField, err)
		}
	}
	t.Log("two saved original records accepted;16 choice,12 return and2 helper-parent contradictions rejected with refreshed self-digests;producer executions0;model0")
}
