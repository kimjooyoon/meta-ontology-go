package languagepackageruntime

import "strings"

func manifestFor(definition Definition) packageruntimeManifest {
	manifest := baseManifest()
	switch definition.Assertion {
	case "PACKAGE_PERMUTATION":
		reversePackages(manifest.Packages)
	case "IMPORT_PERMUTATION":
		reverseStrings(manifest.Packages[3].Imports)
	case "SOURCE_PERMUTATION":
		reverseSources(manifest.Packages[3].Sources)
	}
	switch definition.Mutation {
	case "UNKNOWN_SCHEMA":
		manifest.Schema = "gooo/package-runtime-manifest/unknown"
	case "DUPLICATE_PACKAGE":
		manifest.Packages = append(manifest.Packages, manifest.Packages[0])
	case "UNKNOWN_IMPORT":
		manifest.Packages[3].Imports = append(manifest.Packages[3].Imports, "example/missing")
	case "IMPORT_CYCLE":
		manifest.Packages[0].Imports = []string{"example/app"}
	case "HEADER_MISMATCH":
		manifest.Packages[3].Sources[0].Content = strings.Replace(
			manifest.Packages[3].Sources[0].Content, "package app", "package wrong", 1)
	case "PARSE_ERROR":
		manifest.Packages[3].Sources[0].Content += "\n@"
	case "UNKNOWN_ENTRY_PACKAGE":
		manifest.Entry.PackagePath = "example/missing"
	case "UNKNOWN_ENTRY_ACTIVITY":
		manifest.Entry.Activity = "Missing"
	}
	return manifest
}
