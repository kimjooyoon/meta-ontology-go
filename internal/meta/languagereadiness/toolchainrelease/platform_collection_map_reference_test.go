package toolchainrelease

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectionMapIndependentReference(t *testing.T) {
	values := []string{"", "a", "a-1", "a", "../a", "한글", "A", "a\n", "a\x00", strings.Repeat("a", 64), strings.Repeat("a", 65)}
	raw, err := json.Marshal(map[string]any{"values": values})
	if err != nil {
		t.Fatal(err)
	}
	got, err := collectionMapReference(raw)
	want := `{"values":[false,true,true,true,false,false,false,false,false,true,false]}`
	if err != nil || string(got) != want {
		t.Fatalf("reference = %s, %v; want %s", got, err, want)
	}
	for _, size := range []int{0, 256, 257} {
		raw, _ := json.Marshal(map[string]any{"values": make([]string, size)})
		got, err := collectionMapReference(raw)
		if (err == nil) != (size <= 256) {
			t.Fatalf("size %d: %s, %v", size, got, err)
		}
		if size == 0 && string(got) != `{"values":[]}` {
			t.Fatal("empty collection lost array identity", string(got))
		}
	}
	for _, raw := range []string{`null`, `{}`, `{"values":null}`, `{"values":[null]}`, `{"values":[1]}`,
		`{"values":[true]}`, `{"values":"a"}`, `{"values":[],"extra":1}`} {
		if _, err := collectionMapReference([]byte(raw)); err == nil {
			t.Fatal("invalid reference input accepted", raw)
		}
	}
}

func TestNativeCollectionMapReleaseProfile(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	output := os.Getenv("GOOO_COLLECTION_MAP_SMOKE_OUTPUT")
	if output == "" {
		output = t.TempDir()
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, profile := range computationProfiles {
		if profile.Name == "collection-map" {
			if err := smokeComputationPackage(binary, BuildInput{Root: root, OutputDir: output, Target: Target{ID: "host"}}, profile); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatal("collection map release profile is absent")
}
