package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestObserverAllowsDisjointPathsAndTempSymlinkIdentity(t *testing.T) {
	root := t.TempDir()
	tempRoot := filepath.Join(root, "temp")
	if err := os.Mkdir(tempRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source.gooo")
	output := filepath.Join(root, "output.go")
	if err := os.WriteFile(source, []byte("stable source"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("stable output"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tempRoot, "target.tmp")
	link := filepath.Join(tempRoot, "stable.link")
	if err := os.WriteFile(target, []byte("stable temp"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target.tmp", link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	request := sampleRequest(StatusFail)
	observer, err := NewNoWriteObserver(requestObservationBinding(request), ObserverPaths{
		SourcePath: filepath.Join(root, ".", "source.gooo"),
		OutputPath: filepath.Join(root, "nested", "..", "output.go"),
		TempRoot:   tempRoot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(observer.paths.SourcePath) ||
		!filepath.IsAbs(observer.paths.OutputPath) || !filepath.IsAbs(observer.paths.TempRoot) {
		t.Fatalf("observer paths were not canonical absolute paths: %+v", observer.paths)
	}
	foundSymlink := false
	for _, entry := range observer.before.Temp.Entries {
		if entry.Path == "stable.link" && entry.Kind != "symlink" {
			t.Fatalf("temp symlink lost identity: %+v", entry)
		}
		if entry.Path == "stable.link" {
			foundSymlink = true
		}
	}
	if !foundSymlink {
		t.Fatal("temp symlink was omitted from observer snapshot")
	}
}
