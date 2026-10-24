package packageruntime

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func splitFixture() Manifest {
	source := "// retained header\r\npackage ledger\r\nnamespace ledger\r\n" +
		"entity Integer id \"urn:gooo:type:integer\"\r\n" +
		"activity 보정(Integer) -> Integer id \"urn:split:adjust\" computes `\r\nreturn input+17\r\n`\r\n" +
		"// retained comment\r\nactivity Total(Integer) -> Integer id \"urn:split:total\" computes \"return input-3\"\r\n" +
		"bind 보정.result -> Total.input\r\n// eof retained"
	return Manifest{Schema: ManifestSchema, Entry: EntrySpec{PackagePath: "split/ledger", Activity: "Total"},
		Packages: []PackageSpec{{Path: "split/ledger", Name: "ledger", Sources: []Source{{Filename: "source.gooo", Content: source}}}}}
}

func TestSplitSourceRetainsExactBytesOriginsAndIdentity(t *testing.T) {
	manifest := splitFixture()
	input, _ := json.Marshal(manifest)
	result, err := SplitSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "source.gooo.part02.gooo")
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(manifest)
	if string(input) != string(after) || len(result.Files) != 2 || result.InputImageDigest == result.OutputImageDigest {
		t.Fatal("input mutated or source relocation lost its identity")
	}
	original := manifest.Packages[0].Sources[0].Content
	seen := make([]int, len(original))
	for _, file := range result.Files {
		end := 0
		for _, origin := range file.Origins {
			if origin.OutputStart != end || origin.InputStart < 0 || origin.InputEnd > len(original) ||
				original[origin.InputStart:origin.InputEnd] != file.Source[origin.OutputStart:origin.OutputEnd] {
				t.Fatal("output is not a complete exact source projection", file.Filename, origin)
			}
			for index := origin.InputStart; index < origin.InputEnd; index++ {
				seen[index]++
			}
			end = origin.OutputEnd
		}
		if end != len(file.Source) || file.SHA256 != splitSourceDigest(file.Source) {
			t.Fatal("unmapped output or stale digest", file.Filename)
		}
	}
	headerEnd := result.Files[1].Origins[0].InputEnd
	for index, count := range seen {
		want := 1
		if index < headerEnd {
			want = 2
		}
		if count != want {
			t.Fatalf("source byte %d copied %d times, want %d", index, count, want)
		}
	}
	replay, err := SplitSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "source.gooo.part02.gooo")
	if err != nil || !reflect.DeepEqual(result, replay) {
		t.Fatal("nondeterministic split", err)
	}
}

func TestSplitSourceRejectsMissingBoundaryAndOccupiedDestination(t *testing.T) {
	for _, row := range []struct{ id, into string }{
		{"urn:absent", "new.gooo"}, {"urn:gooo:type:integer", "new.gooo"},
		{"urn:split:total", "SOURCE.gooo"}, {"urn:split:total", "../new.gooo"},
		{"urn:split:total", "new.go"}, {"urn:split:total", "nested/../new.gooo"},
	} {
		manifest := splitFixture()
		before, _ := json.Marshal(manifest)
		result, err := SplitSource(manifest, "split/ledger", "source.gooo", row.id, row.into)
		after, _ := json.Marshal(manifest)
		if err == nil || len(result.Files) != 0 || string(before) != string(after) {
			t.Fatal(row, result, err)
		}
	}
}

func TestSplitSourceRetainsFileLocalImportContext(t *testing.T) {
	manifest := splitFixture()
	manifest.Packages = append(manifest.Packages, PackageSpec{Path: "split/math", Name: "math", Sources: []Source{
		{Filename: "math.gooo", Content: "package math\nnamespace math\nentity Integer id \"urn:gooo:type:integer\"\nactivity Identity(Integer) -> Integer computes \"return input\"\n"},
	}})
	manifest.Packages[0].Imports = []string{"split/math"}
	source := &manifest.Packages[0].Sources[0].Content
	*source = strings.Replace(*source, "namespace ledger\r\n", "namespace ledger\r\nimport math \"split/math\"\r\n", 1)
	*source = strings.Replace(*source, "Total(Integer) -> Integer", "Total(math.Integer) -> math.Integer", 1)
	result, err := SplitSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "second.gooo")
	if err != nil || !strings.Contains(result.Files[1].Source, "import math \"split/math\"") {
		t.Fatal(result, err)
	}
	*source = strings.Replace(*source, "import math \"split/math\"\r\n", "", 1)
	*source += "\r\nimport math \"split/math\"\r\n"
	if result, err := SplitSource(manifest, "split/ledger", "source.gooo", "urn:split:total", "second.gooo"); err == nil || len(result.Files) != 0 {
		t.Fatal("interleaved import copied with a different scope", result)
	}
}
