package languagepackageruntime

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"

func measureRuntime(result *CaseResult, runtime packageruntime.Result) {
	result.Packages = len(runtime.Image.Packages)
	result.Initializations = len(runtime.Image.InitOrder)
	result.EntryBindings = boolCount(runtime.Image.Entry.Activity != "")
	result.SemanticBindings = semanticBindings(runtime)
	result.Effects, result.RepositoryWrites = runtime.Effects, runtime.RepositoryWrites
	for _, image := range runtime.Image.Packages {
		result.Sources += len(image.Sources)
		result.Imports += len(image.Imports)
	}
}

func semanticBindings(runtime packageruntime.Result) int {
	count := 0
	for _, image := range runtime.Image.Packages {
		for _, source := range image.Sources {
			if source.SemanticDigest != "" {
				count++
			}
		}
	}
	return count
}

func packageSources(runtime packageruntime.Result, packagePath string) int {
	for _, image := range runtime.Image.Packages {
		if image.Path == packagePath {
			return len(image.Sources)
		}
	}
	return 0
}

func isPermutation(assertion string) bool {
	return assertion == "PACKAGE_PERMUTATION" || assertion == "IMPORT_PERMUTATION" ||
		assertion == "SOURCE_PERMUTATION"
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}
