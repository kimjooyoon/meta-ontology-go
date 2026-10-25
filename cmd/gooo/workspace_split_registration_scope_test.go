package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceRegisteredSplitRequiresNativeAdmissionForEmbeddedConsumers(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	for _, kind := range []string{"single", "wildcard", "test", "inactive", "unrelated"} {
		t.Run(kind, func(t *testing.T) {
			options := registeredSplitFixture(t)
			writeRegisteredSplitInput(t, options.registryRoot, "go.mod", []byte("module example.test/split-consumer\n\ngo 1.27.3\n"))
			pattern, name, prefix := splitWorkspaceDirectory+"/main.gooo", "consumer.go", ""
			if kind == "wildcard" {
				pattern = splitWorkspaceDirectory + "/*.gooo"
			}
			if kind == "test" {
				name = "consumer_test.go"
			}
			if kind == "inactive" {
				prefix = "//go:build gooo_not_selected\n\n"
				writeRegisteredSplitInput(t, options.registryRoot, "package.go", []byte("package fixture\n"))
			}
			if kind == "unrelated" {
				pattern = "other.txt"
				writeRegisteredSplitInput(t, options.registryRoot, pattern, []byte("unrelated text"))
			}
			writeRegisteredSplitInput(t, options.registryRoot, name,
				[]byte(prefix+"package fixture\nimport _ \"embed\"\n//go:embed "+pattern+"\nvar source string\n"))
			result, err := prepareWorkspaceSourceSplit(options, OSFileReader{})
			if kind == "unrelated" {
				if err != nil || result.Decision != "PROPOSED" {
					t.Fatal("unrelated embed prevented a source-only split", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "native") || result.Decision == "PROPOSED" || result.Registration != nil || result.Result != nil {
				t.Fatal("embedded runtime consumer was silently treated as admitted", err)
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(options.manifestPath), "inspect.gooo")); !os.IsNotExist(err) {
				t.Fatal("consumer inspection installed a fragment")
			}
		})
	}
}

func TestWorkspaceRegisteredSplitFlagCombinationsAndRoot(t *testing.T) {
	for _, args := range [][]string{
		{"--apply", "saved.json", "--syntax-registry-root", ".", "workspace.json"},
		{"--boundaries", "--package", "p", "--source", "main.gooo", "--syntax-registry-root", ".", "workspace.json"},
	} {
		if _, err := parseWorkspaceSourceSplit(args, &strings.Builder{}); err == nil {
			t.Fatal("invalid registry option combination accepted")
		}
	}
	options := registeredSplitFixture(t)
	options.registryRoot = t.TempDir()
	result, err := prepareWorkspaceSourceSplit(options, OSFileReader{})
	if err == nil || result.Decision == "PROPOSED" {
		t.Fatal("workspace outside registry root accepted")
	}
}

func TestWorkspaceRegisteredSplitAncestorReadFailure(t *testing.T) {
	root := t.TempDir()
	if _, err := captureSplitAncestorGo(root, filepath.Join(root, "missing"), map[string][]byte{}); err == nil {
		t.Fatal("an unreadable ancestor was reported as zero native consumers")
	}
}
