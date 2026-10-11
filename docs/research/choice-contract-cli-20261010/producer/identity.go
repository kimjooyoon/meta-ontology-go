package main

import (
	"crypto/sha256"
	"debug/buildinfo"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var fixtureSHA = map[string]string{
	"fresh-bound-k9-r1-w0-alias-g0.gooo":  "fdf74b1a109574621a381e775d56194877fe3537d117ee4adfe41cfb4504f5a4",
	"fresh-bound-k9-r1-w0-alias-g0.json":  "7dec9a45b303afd5e16a048963237a6bfb5d83731f0d43ed727f39d31709be12",
	"fresh-bound-k9-r1-w0-alias-g1.gooo":  "7ac611dd4fcb2ec1ac1fb21d7530c2aa7f820083bc15881862e8581ca2804ebb",
	"fresh-bound-k9-r1-w0-alias-g1.json":  "b91cdb1bc85ca4dd24f95dc2ee528cbbc25a2c5356a9faf547bb913fa5d25b8f",
	"fresh-scale-k9-r1-w0-direct-g1.gooo": "e0f35a8ca157800aad47bcf7b668a0e73032dd585ddcb02f06773a41a358cad6",
	"fresh-scale-k9-r1-w0-direct-g1.json": "1d989786e0752f2b903dc552c4050757e86a084d43ed9b4f5a57c142e83a9915",
}

func fixedFixture(root, name string) ([]byte, error) {
	want, ok := fixtureSHA[name]
	if !ok {
		return nil, errors.New("unknown fixed fixture")
	}
	raw, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != want {
		return nil, fmt.Errorf("fixed source changed: %s", name)
	}
	return raw, nil
}

func checkCompilerBuild(path string) error {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return err
	}
	if info.GoVersion != "go1.27.2" || info.Path != "github.com/kimjooyoon/meta-ontology-go/cmd/gooo" {
		return errors.New("exact compiler/toolchain required")
	}
	clean, revision, publicSDK := false, "", false
	for _, item := range info.Settings {
		if item.Key == "vcs.modified" {
			clean = item.Value == "false"
		}
		if item.Key == "vcs.revision" {
			revision = item.Value
		}
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/kimjooyoon/gooo-decision-runtime" {
			publicSDK = dep.Version == "v0.2.37-experimental" && dep.Replace == nil
		}
	}
	if !clean || len(revision) != 40 || !publicSDK {
		return errors.New("clean committed compiler and public SDK0.2.37 required")
	}
	return nil
}
