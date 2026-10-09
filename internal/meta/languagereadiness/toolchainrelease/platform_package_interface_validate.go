package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type packageInterfaceSmokeReceipt struct {
	Schema, Decision, Manifest, Error string
	ManifestDigest                    string                    `json:"manifest_digest"`
	Interface                         *packageruntime.Interface `json:"interface"`
}

func validatePackageInterfaceSmoke(raw []byte, root string) error {
	var r packageInterfaceSmokeReceipt
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	manifest, _, err := digestFile(filepath.Join(root, packageInterfaceManifest))
	if err != nil {
		return err
	}
	if r.Schema != "gooo/package-workspace-interface-receipt/v1" || r.Decision != "PASS" || r.Error != "" ||
		r.Manifest != packageInterfaceManifest || r.ManifestDigest != manifest || r.Interface == nil {
		return fmt.Errorf("package interface envelope or source manifest differs")
	}
	view := *r.Interface
	digest := view.Digest
	view.Digest = ""
	bound, err := digestValue(view)
	if err != nil {
		return err
	}
	if digest != bound || view.Schema != "gooo/package-interface/v1" ||
		view.Scope != "DECLARED_PACKAGE_INTERFACES" || view.ImageDigest == "" ||
		view.Entry != (packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"}) {
		return fmt.Errorf("package interface identity, entry or scope differs")
	}
	return validateInterfacePackages(view.Packages, root)
}

func validateInterfacePackages(packages []packageruntime.InterfacePackage, root string) error {
	want := expectedInterfacePackages()
	if len(packages) != len(want) {
		return fmt.Errorf("package interface package count differs")
	}
	for i, pkg := range packages {
		filename := want[i].Name + ".gooo.fixture"
		raw, err := os.ReadFile(filepath.Join(root, filepath.Dir(packageInterfaceManifest), filename))
		if err != nil {
			return err
		}
		// Package source identities hash the JSON-encoded source string.
		source, err := digestValue(string(raw))
		if err != nil {
			return err
		}
		if len(pkg.Sources) != 1 || pkg.Sources[0].Filename != filename ||
			pkg.Sources[0].SourceDigest != source || pkg.Sources[0].SemanticDigest == "" ||
			pkg.Sources[0].Declarations != len(want[i].Declarations) {
			return fmt.Errorf("package interface source ownership differs for %s", want[i].Path)
		}
		pkg.Sources = nil
		if !reflect.DeepEqual(pkg, want[i]) {
			return fmt.Errorf("package interface declarations differ for %s", want[i].Path)
		}
	}
	return nil
}

func expectedInterfacePackages() []packageruntime.InterfacePackage {
	const profile = "example://domain/profile"
	activity := func(name, id, source string) packageruntime.InterfaceDeclaration {
		return packageruntime.InterfaceDeclaration{Kind: "activity", Name: name, ID: id,
			Source: source, Inputs: []string{profile}, Output: profile}
	}
	fields := make([]packageruntime.InterfaceField, 0, 3)
	for _, pair := range [][2]string{{"note", "string"}, {"complete", "boolean"}, {"count", "integer"}} {
		fields = append(fields, packageruntime.InterfaceField{Name: pair[0], ID: profile + "/" + pair[0],
			TypeID: "urn:gooo:type:" + pair[1], Presence: "optional", Cardinality: "one"})
	}
	return []packageruntime.InterfacePackage{
		{Path: "example/domain", Name: "domain", Namespace: "domain", Declarations: []packageruntime.InterfaceDeclaration{
			activity("Submit", "domain://activity/submit", "domain.gooo.fixture"),
			{Kind: "entity", Name: "Profile", ID: profile, Source: "domain.gooo.fixture", Shape: "record", Fields: fields},
		}},
		{Path: "example/app", Name: "app", Namespace: "app", Imports: []string{"example/domain"},
			Declarations: []packageruntime.InterfaceDeclaration{activity("Main", "app://activity/main", "app.gooo.fixture")}},
	}
}
