package embeduse_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle/embeduse"
)

func TestGoResolvesDirectoryAllAndQuotedPatterns(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string]string{
		"go.mod": "module example.test/patterns\n\ngo 1.27.3\n",
		"consumer.go": `package p
import "embed"
var (
 //go:embed all:assets
 complete embed.FS
 //go:embed assets
 visible embed.FS
 //go:embed "quoted source.gooo"
 quoted []byte
)
`,
		"assets/.hidden.gooo": first, "assets/visible.gooo": first, "quoted source.gooo": first,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	report, err := embeduse.Inspect(context.Background(), root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{"complete": {"assets/.hidden.gooo", "assets/visible.gooo"},
		"visible": {"assets/visible.gooo"}, "quoted": {"quoted source.gooo"}}
	for _, consumer := range report.Consumers {
		var names []string
		for _, index := range consumer.Members {
			names = append(names, report.Files[index].Filename)
		}
		if !reflect.DeepEqual(names, want[consumer.Name]) {
			t.Fatalf("Go pattern result changed: %s %v", consumer.Name, names)
		}
		delete(want, consumer.Name)
	}
	if len(want) != 0 {
		t.Fatal("variable block lost consumers")
	}
}

func TestObservationRemovesOwnedQueryFiles(t *testing.T) {
	root := fixture(t)
	temporary := t.TempDir()
	t.Setenv("TMPDIR", temporary)
	t.Setenv("TMP", temporary)
	t.Setenv("TEMP", temporary)
	if _, err := embeduse.Inspect(context.Background(), root, partition()); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "contract.gooo")); err != nil {
		t.Fatal(err)
	}
	if _, err := embeduse.Inspect(context.Background(), root, nil); err == nil {
		t.Fatal("missing input unexpectedly succeeded")
	}
	entries, err := os.ReadDir(temporary)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "gooo-embed-use-") {
			t.Fatalf("owned query files leaked: %s", entry.Name())
		}
	}
}
