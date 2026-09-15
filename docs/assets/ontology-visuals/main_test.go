package main

import "testing"

func TestVisualContract(t *testing.T) {
	manifest, err := loadManifest("visual-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateManifest(manifest); err != nil {
		t.Fatal(err)
	}
	if err := checkAssets(".", manifest); err != nil {
		t.Fatal(err)
	}
}
