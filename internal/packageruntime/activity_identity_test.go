package packageruntime

import (
	"strings"
	"testing"
)

func TestExplicitActivityIdentityIsExportedAcrossRename(t *testing.T) {
	for _, name := range []string{"Choose", "Pick", "Canonical"} {
		m := interfaceFixture()
		m.Entry.Activity = name
		m.Packages[0].Sources[0].Content = strings.ReplaceAll(m.Packages[0].Sources[0].Content, "Choose", name)
		id := "urn:gooo:activity:choose"
		if name == "Canonical" {
			id = "URN:gooo:activity:choose"
		}
		m.Packages[0].Sources[0].Content = strings.ReplaceAll(m.Packages[0].Sources[0].Content, " computes", ` id "`+id+`" computes`)
		view, err := Describe(m)
		if err != nil {
			t.Fatal(err)
		}
		a := view.Packages[1].Declarations[0]
		if a.ID != "urn:gooo:activity:choose" || a.Name != name || len(a.Inputs) != 3 || a.Output != "example://profile" {
			t.Fatal("interface lost explicit identity or signature", a)
		}
		image, err := Build(m)
		if err != nil || image.Packages[1].Exports[0].ID != "urn:gooo:activity:choose" {
			t.Fatal("package export lost explicit identity", image, err)
		}
	}
}

func TestBuildRejectsActivityIdentityCollisionAcrossSources(t *testing.T) {
	for _, id := range []string{"urn:gooo:activity:choose", "app://activity/choose"} {
		m := interfaceFixture()
		if strings.HasPrefix(id, "urn:") {
			m.Packages[0].Sources[0].Content = strings.ReplaceAll(m.Packages[0].Sources[0].Content, " computes", ` id "`+id+`" computes`)
		}
		m.Packages[0].Sources = append(m.Packages[0].Sources, Source{Filename: "other.gooo", Content: "package app\nnamespace app\nactivity Other(Profile, Text, Profile) -> Profile id \"" + id + "\" computes \"return input0\""})
		image, err := Build(m)
		if err == nil || image.Schema != "" {
			t.Fatal("cross-source identity collision accepted", image, err)
		}
	}
}
