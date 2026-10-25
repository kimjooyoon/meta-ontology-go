package embeduse_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle/embeduse"
)

const header = "package contract\nnamespace contract\n"
const first = header + "entity Input id \"gooo://contract/input\"\n"
const tail = "entity Result id \"gooo://contract/result\"\nactivity Transform(Input) -> Result\n"
const consumerSource = `package p
import "embed"
//line synthetic.go:900
//go:embed contract*.gooo
var family embed.FS
//go:embed contract.gooo
var single []byte
`

func fixture(t *testing.T) string {
	t.Helper()
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	root := t.TempDir()
	authored, err := os.ReadFile(filepath.Join("testdata", "contract.gooo.fixture"))
	if err != nil || string(authored) != first+tail {
		t.Fatalf("authored contract and independent partition expectations differ: %v", err)
	}
	for name, raw := range map[string]string{
		"go.mod":      "module example.test/embed-origin\n\ngo 1.27.3\n",
		"consumer.go": consumerSource, "contract.gooo": string(authored),
		"inside_test.go":  "package p\nimport _ \"embed\"\n//go:embed contract.gooo\nvar inside []byte\n",
		"outside_test.go": "package p_test\nimport \"embed\"\n//go:embed contract*.gooo\nvar outside embed.FS\n",
		"inactive.go":     "//go:build gooo_alt\n\npackage p\nimport _ \"embed\"\n//go:embed contract.gooo\nvar alternative []byte\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func partition() []sourcebundle.Source {
	return []sourcebundle.Source{{Filename: "contract.gooo", Content: first},
		{Filename: "contract-extra.gooo", Content: header + tail}}
}

func TestProjectedMembersStayBoundToEachGoVariable(t *testing.T) {
	root := fixture(t)
	before, err := embeduse.Inspect(context.Background(), root, nil)
	if err != nil {
		t.Fatal(err)
	}
	after, err := embeduse.Inspect(context.Background(), root, partition())
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Consumers) != 4 || len(after.Consumers) != 4 || !reflect.DeepEqual(after.IgnoredGoFiles, []string{"inactive.go"}) {
		t.Fatalf("selected contexts lost: %+v", after)
	}
	want := map[string]int{"family": 2, "single": 1, "inside": 1, "outside": 2}
	if len(after.Projection) != 2 || len(before.Projection) != 0 {
		t.Fatal("projection input set was lost")
	}
	for index, consumer := range after.Consumers {
		if len(before.Consumers[index].Members) != 1 || len(consumer.Members) != want[consumer.Name] {
			t.Fatalf("package union hid a variable's actual input: %+v", consumer)
		}
		for _, member := range consumer.Members {
			if !after.Files[member].Projected {
				t.Fatal("projected bytes reported as observed disk bytes")
			}
		}
		original, err := os.ReadFile(filepath.Join(root, consumer.Declaration.File))
		if err != nil {
			t.Fatal(err)
		}
		file := after.Files[consumer.GoSource]
		if file.Projected || file.SHA256 != fmt.Sprintf("%x", sha256.Sum256(original)) ||
			!strings.Contains(string(original[consumer.Declaration.Start:consumer.Declaration.End]), "var "+consumer.Name) {
			t.Fatalf("consumer origin was replaced by an isolated tool-query source: %+v", consumer)
		}
		if consumer.Name == "family" && (consumer.Declaration.StartLine != 4 || consumer.Declaration.File != "consumer.go") {
			t.Fatal("Go line directive replaced the physical source origin")
		}
	}
	if after.Build["GOVERSION"] == "" || after.Build["GOOS"] == "" || after.Build["GOARCH"] == "" {
		t.Fatal("build context missing")
	}
	assertOriginalFiles(t, root)
	replay, err := embeduse.Inspect(context.Background(), root, partition())
	if err != nil || !reflect.DeepEqual(after, replay) {
		t.Fatalf("observation replay changed: %v", err)
	}
	assertPartitionMeaning(t)
}

func assertOriginalFiles(t *testing.T, root string) {
	t.Helper()
	for name, want := range map[string]string{"consumer.go": consumerSource, "contract.gooo": first + tail} {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(raw) != want {
			t.Fatalf("source %s was changed: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "contract-extra.gooo")); !os.IsNotExist(err) {
		t.Fatal("projected member was installed")
	}
}

func assertPartitionMeaning(t *testing.T) {
	t.Helper()
	before, err := sourcebundle.Lower([]sourcebundle.Source{{Filename: "contract.gooo", Content: first + tail}})
	if err != nil {
		t.Fatal(err)
	}
	after, err := sourcebundle.Lower(partition())
	if err != nil {
		t.Fatal(err)
	}
	partial, err := sourcebundle.Lower(partition()[:1])
	if err != nil {
		t.Fatal(err)
	}
	if before.IR.StableHash() != after.IR.StableHash() || before.IR.StableHash() == partial.IR.StableHash() ||
		len(before.Provenance.Declarations) != 3 || len(after.Provenance.Declarations) != 3 || len(partial.Provenance.Declarations) != 1 {
		t.Fatal("independent Gooo reader did not expose the single consumer's missing declarations")
	}
}
