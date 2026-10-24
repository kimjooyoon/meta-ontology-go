package moduleboundary_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// Exercise the public planner command, including its fixed-point shortcut.
// Compare the whole disposable tree so a late rejection cannot hide early writes.
func TestPackagePartitionModuleContract(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "planner")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.CommandContext(t.Context(), filepath.Join(runtime.GOROOT(), "bin", "go"),
		"build", "-o", binary, "./bootstrap/logical-split-planner")
	build.Dir = filepath.Join("..", "..", "..")
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build planner: %v\n%s", err, output)
	}
	cases := []struct {
		name  string
		edit  func(*testing.T, string, map[string]any)
		owner string
	}{
		{"nested module", nil, "nested"},
		{"root module", func(t *testing.T, root string, _ map[string]any) {
			if err := os.Remove(filepath.Join(root, "nested", "go.mod")); err != nil {
				t.Fatal(err)
			}
		}, "."},
		{"cross-module move", func(_ *testing.T, _ string, r map[string]any) {
			r["moves"] = []map[string]any{{"Source": "nested/pkg/source.go", "Destination": "outside/new/source.go"}}
		}, ""},
		{"cross-module creation", func(_ *testing.T, _ string, r map[string]any) {
			r["creates"] = []map[string]any{{"Path": "outside/new/source.go", "Content": "package outside\n"}}
		}, ""},
		{"cross-module rewrite", func(_ *testing.T, _ string, r map[string]any) {
			r["rewrites"] = []map[string]any{{"Path": "outside/readme.txt", "Old": "original", "New": "changed", "expected_count": 1}}
		}, ""},
		{"cross-module range", func(_ *testing.T, _ string, r map[string]any) {
			r["ranges"] = []map[string]any{{"Path": "outside/readme.txt", "Start": "original", "Replacement": "changed"}}
		}, ""},
		{"new module declaration", func(_ *testing.T, _ string, r map[string]any) {
			r["creates"] = []map[string]any{{"Path": "nested/pkg/command/go.mod", "Content": "module example.test/new\n"}}
		}, ""},
		{"module declaration rewrite", func(_ *testing.T, _ string, r map[string]any) {
			r["rewrites"] = []map[string]any{{"Path": "nested/go.mod", "Old": "example.test/nested", "New": "example.test/replaced", "expected_count": 1}}
		}, ""},
		{"module declaration alias", func(t *testing.T, root string, r map[string]any) {
			if err := os.Symlink("go.mod", filepath.Join(root, "nested", "module-link")); err != nil {
				t.Skip(err)
			}
			r["rewrites"] = []map[string]any{{"Path": "nested/module-link", "Old": "example.test/nested", "New": "example.test/replaced", "expected_count": 1}}
		}, ""},
		{"cross-module directory alias", func(t *testing.T, root string, r map[string]any) {
			if err := os.Symlink("../../outside", filepath.Join(root, "nested", "pkg", "alias")); err != nil {
				t.Skip(err)
			}
			r["moves"] = []map[string]any{{"Source": "nested/pkg/source.go", "Destination": "nested/pkg/alias/new/source.go"}}
		}, ""},
		{"unavailable toolchain", func(t *testing.T, root string, _ map[string]any) {
			writePartitionFile(t, root, "nested/go.mod", "module example.test/nested\ngo 1.99.0\n")
		}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			work := t.TempDir()
			root := filepath.Join(work, "repo")
			for path, data := range map[string]string{
				"go.mod": "module example.test/root\n", "nested/go.mod": "module example.test/nested\n",
				"nested/pkg/source.go": "package main\nfunc Value() int { return 1 }\n",
				"outside/readme.txt":   "original\n",
			} {
				writePartitionFile(t, root, path, data)
			}
			recipe := map[string]any{
				"schema": "gooo.go-package-partition-recipe.v1", "subject": "nested/pkg",
				"moves":          []map[string]any{{"Source": "nested/pkg/source.go", "Destination": "nested/pkg/command/source.go"}},
				"creates":        []map[string]any{{"Path": "nested/pkg/helper/helper.go", "Content": "package helper\n"}},
				"expected_shape": map[string]any{"branch_entries": 2, "max_entries": 10, "leaves": map[string]int{"command": 1, "helper": 1}},
			}
			if tc.edit != nil {
				tc.edit(t, root, recipe)
			}
			payload, err := json.Marshal(recipe)
			if err != nil {
				t.Fatal(err)
			}
			recipePath := filepath.Join(work, "recipe.json")
			if err := os.WriteFile(recipePath, payload, 0o600); err != nil {
				t.Fatal(err)
			}
			receiptPath := filepath.Join(work, "receipt.json")
			before := partitionSnapshot(t, root)
			run := func() ([]byte, error) {
				cmd := exec.CommandContext(t.Context(), binary, "--root", root, "--package-partition-recipe", recipePath,
					"--expected-sha", "fixture-source", "--output", receiptPath)
				cmd.Env = append(os.Environ(), "GOOO_DISPOSABLE_ROOT="+root)
				return cmd.CombinedOutput()
			}
			output, err := run()
			if tc.owner == "" {
				if err == nil {
					t.Fatalf("unbound partition accepted: %s", output)
				}
				if !reflect.DeepEqual(before, partitionSnapshot(t, root)) {
					t.Fatal("module rejection left a partially modified tree")
				}
				if _, err := os.Stat(receiptPath); !os.IsNotExist(err) {
					t.Fatalf("rejected plan published a receipt: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("valid partition: %v %s", err, output)
			}
			assertPartitionReceipt(t, receiptPath, "PASS", tc.owner)
			compile := exec.CommandContext(t.Context(), filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "./...")
			compile.Dir = filepath.Join(root, tc.owner)
			compile.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOPROXY=off")
			if output, err := compile.CombinedOutput(); err != nil {
				t.Fatalf("partition did not compile in its owning module: %v %s", err, output)
			}
			after := partitionSnapshot(t, root)
			if output, err := run(); err != nil {
				t.Fatalf("fixed point: %v %s", err, output)
			}
			assertPartitionReceipt(t, receiptPath, "FIXED_POINT", tc.owner)
			if !reflect.DeepEqual(after, partitionSnapshot(t, root)) {
				t.Fatal("fixed point modified source")
			}
			writePartitionFile(t, root, filepath.Join(tc.owner, "go.mod"), "module example.test/changed\ngo 1.99.0\n")
			if output, err := run(); err == nil {
				t.Fatalf("fixed point bypassed changed module requirement: %s", output)
			}
		})
	}
}

func writePartitionFile(t *testing.T, root, path, data string) {
	t.Helper()
	name := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func partitionSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		value := "directory"
		if entry.Type()&os.ModeSymlink != 0 {
			value, err = os.Readlink(path)
		} else if !entry.IsDir() {
			var data []byte
			data, err = os.ReadFile(path)
			value = string(data)
		}
		out[path] = fmt.Sprintf("%v:%s", entry.Type(), value)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func assertPartitionReceipt(t *testing.T, path, decision, owner string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var receipt map[string]any
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt["decision"] != decision || receipt["module_root"] != owner {
		t.Fatalf("partition lost module origin: %s", data)
	}
}
