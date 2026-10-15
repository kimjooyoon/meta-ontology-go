package main

import "fmt"

func selectStarterFiles(template, module string, moduleProvided bool) (map[string]string, string, error) {
	switch template {
	case "app":
		if moduleProvided {
			return nil, "", fmt.Errorf("--module is available only with --template library, diagnostic or origins")
		}
		return withStarterGuide(starterFiles), module, nil
	case "library", "diagnostic", "origins":
		if (template == "diagnostic" || template == "origins") && !moduleProvided {
			module = template
		}
		if !libraryModulePathPattern.MatchString(module) {
			return nil, "", fmt.Errorf("invalid %s module path %q", template, module)
		}
		if template == "origins" {
			return withStarterGuide(originsStarterFiles), module, nil
		}
		if template == "diagnostic" {
			return withStarterGuide(diagnosticStarterFiles), module, nil
		}
		return withStarterGuide(libraryStarterFiles), module, nil
	default:
		return nil, "", fmt.Errorf("unknown template %q (choose app, library, diagnostic or origins)", template)
	}
}
