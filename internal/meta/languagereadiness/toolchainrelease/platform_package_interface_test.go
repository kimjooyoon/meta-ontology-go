package toolchainrelease

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func packageInterfaceFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/package-interface.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestPackageInterfaceSmokeRequiresSourceOwnedFields(t *testing.T) {
	raw := packageInterfaceFixture(t)
	if err := validatePackageInterfaceSmoke(raw, "../../../.."); err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]nativeSmokeEdit{
		"schema":             {[]any{"schema"}, "changed"},
		"decision":           {[]any{"decision"}, "FAIL_CLOSED"},
		"error":              {[]any{"error"}, "failed"},
		"manifest":           {[]any{"manifest"}, "other.json"},
		"manifest digest":    {[]any{"manifest_digest"}, "sha256:changed"},
		"missing interface":  {[]any{"interface"}, nil},
		"scope":              {[]any{"interface", "scope"}, "changed"},
		"entry":              {[]any{"interface", "entry", "activity"}, "Other"},
		"image":              {[]any{"interface", "image_digest"}, ""},
		"package ownership":  {[]any{"interface", "packages", 0, "path"}, "example/app"},
		"source":             {[]any{"interface", "packages", 0, "sources", 0, "source_digest"}, "sha256:changed"},
		"declaration count":  {[]any{"interface", "packages", 0, "sources", 0, "declarations"}, json.Number("3")},
		"import placeholder": {[]any{"interface", "packages", 1, "declarations", 0, "kind"}, "entity"},
		"input order/count":  {[]any{"interface", "packages", 0, "declarations", 0, "inputs"}, []string{}},
		"record shape":       {[]any{"interface", "packages", 0, "declarations", 1, "shape"}, "nominal"},
		"field type":         {[]any{"interface", "packages", 0, "declarations", 1, "fields", 2, "type_id"}, "urn:gooo:type:string"},
		"field presence":     {[]any{"interface", "packages", 0, "declarations", 1, "fields", 2, "presence"}, "required"},
		"field cardinality":  {[]any{"interface", "packages", 0, "declarations", 1, "fields", 2, "cardinality"}, "many"},
	} {
		t.Run(name, func(t *testing.T) {
			changed := changedNativeSmoke(t, raw, edit)
			// Rebind the digest so field/source checks must reject the changed facts.
			var receipt packageInterfaceSmokeReceipt
			if err := json.Unmarshal(changed, &receipt); err != nil {
				t.Fatal(err)
			}
			if receipt.Interface != nil {
				receipt.Interface.Digest = ""
				receipt.Interface.Digest, _ = digestValue(receipt.Interface)
			}
			changed, _ = json.Marshal(receipt)
			if err := validatePackageInterfaceSmoke(changed, "../../../.."); err == nil {
				t.Fatal("changed package interface accepted")
			}
		})
	}
	for _, bad := range [][]byte{nil, []byte(`{}`), append(bytes.Clone(raw), []byte(`{}`)...),
		changedNativeSmoke(t, raw, nativeSmokeEdit{[]any{"interface", "digest"}, "sha256:changed"})} {
		if err := validatePackageInterfaceSmoke(bad, "../../../.."); err == nil {
			t.Fatal("incomplete or unbound interface accepted")
		}
	}
}

func TestPackageInterfaceSmokeRetainsEachOutputAndOriginalError(t *testing.T) {
	for _, output := range [][]byte{nil, []byte("partial JSON\noriginal error\n")} {
		input := BuildInput{Root: "../../../..", OutputDir: t.TempDir(), Target: Target{ID: "host"}}
		failure := errors.New("original process error")
		run := func(string, []string, string, ...string) ([]byte, error) { return output, failure }
		if err := smokePackageInterfaceWithRunner("compiler", input, run); !errors.Is(err, failure) {
			t.Fatal("original process error lost", err)
		}
		retained, err := os.ReadFile(filepath.Join(input.OutputDir, "host-package-interface-initial.failed-output"))
		if err != nil || !bytes.Equal(retained, output) {
			t.Fatal("original output lost", err)
		}
	}
	input := BuildInput{Root: "../../../..", OutputDir: t.TempDir(), Target: Target{ID: "host"}}
	calls := 0
	run := func(root string, env []string, binary string, args ...string) ([]byte, error) {
		calls++
		want := []string{"package", "interface", "--json", packageInterfaceManifest}
		if calls == 3 {
			want = []string{"package", "interface", packageInterfaceManifest}
		}
		if root != input.Root || len(env) != 0 || binary != "compiler" || !slices.Equal(args, want) {
			t.Fatal("interface command changed", root, env, binary, args)
		}
		if calls == 3 {
			return []byte(packageInterfacePlain), nil
		}
		return packageInterfaceFixture(t), nil
	}
	if err := smokePackageInterfaceWithRunner("compiler", input, run); err != nil || calls != 3 {
		t.Fatal("interface profile did not finish", calls, err)
	}
}

func TestNativePackageInterfaceReleaseProfile(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	if err := smokePackageInterface(binary, BuildInput{Root: root, OutputDir: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
}

func TestPackageInterfaceSmokeRejectsChangedRepeatAndPlainOutput(t *testing.T) {
	for _, failAt := range []int{2, 3} {
		input := BuildInput{Root: "../../../..", OutputDir: t.TempDir(), Target: Target{ID: "host"}}
		calls := 0
		run := func(string, []string, string, ...string) ([]byte, error) {
			calls++
			raw := packageInterfaceFixture(t)
			if calls == 3 {
				raw = []byte(packageInterfacePlain)
			}
			if calls == failAt {
				raw = append(raw, '\n')
			}
			return raw, nil
		}
		if err := smokePackageInterfaceWithRunner("compiler", input, run); err == nil || calls != failAt {
			t.Fatal("changed output accepted", failAt, calls, err)
		}
	}
}
