package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceSourceSplitProjectsDirectoryEntriesBeforePublish(t *testing.T) {
	for _, mode := range []string{"directory", "case", "hardlink", "copy"} {
		t.Run(mode, func(t *testing.T) {
			options, _, inputs := sharedApplicationFixture(t)
			root := filepath.Dir(options.manifestPath)
			source := createSplitAlias(t, root, mode, inputs["main.gooo"])
			saveSplitAliasProposal(t, options, inputs, source)
			applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
			if mode == "directory" || mode == "case" {
				if err == nil || applied.Decision == "APPLIED" || applied.Application.State != "PRE_STATE_RETAINED" {
					t.Fatal("unplanned alias effect was published", err, applied)
				}
				checkSourceSplitApplicationInputs(t, options, inputs)
				return
			}
			if err != nil || applied.Decision != "APPLIED" {
				t.Fatal("independent write occurrence was coupled", err, applied)
			}
			if err := verifyAppliedSourceSplit(applied); err != nil {
				t.Fatal(err)
			}
			unchanged, err := os.ReadFile(filepath.Join(root, source))
			if err != nil || !bytes.Equal(unchanged, inputs["main.gooo"]) {
				t.Fatal("independent source changed", err)
			}
		})
	}
}

func TestWorkspaceSourceSplitRechecksAliasAfterStaging(t *testing.T) {
	options, _, inputs := sharedApplicationFixture(t)
	root := filepath.Dir(options.manifestPath)
	isolated := t.TempDir()
	if err := os.WriteFile(filepath.Join(isolated, "main.gooo"), inputs["main.gooo"], 0o640); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(isolated, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	saveSplitAliasProposal(t, options, inputs, "alias/main.gooo")
	ops := defaultAtomicFileOps()
	ops.beforePublish = func() error {
		if err := os.Remove(alias); err != nil {
			return err
		}
		return os.Symlink(root, alias)
	}
	applied, err := applyWorkspaceSourceSplit(options, ops)
	if err == nil || applied.Application.State != "PRE_STATE_RETAINED" {
		t.Fatal("same-byte alias retarget reached publication", err, applied)
	}
	checkSourceSplitApplicationInputs(t, options, inputs)
}
