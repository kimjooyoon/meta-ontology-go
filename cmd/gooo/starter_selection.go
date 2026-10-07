package main

import "fmt"

func selectStarterFiles(template, module string, moduleProvided bool) (map[string]string, string, error) {
	switch template {
	case "app":
		if moduleProvided {
			return nil, "", fmt.Errorf("--module is available only with --template library or diagnostic")
		}
		return starterFiles, module, nil
	case "library", "diagnostic":
		if template == "diagnostic" && !moduleProvided {
			module = "diagnostic"
		}
		if !libraryModulePathPattern.MatchString(module) {
			return nil, "", fmt.Errorf("invalid %s module path %q", template, module)
		}
		if template == "diagnostic" {
			return diagnosticStarterFiles, module, nil
		}
		return libraryStarterFiles, module, nil
	default:
		return nil, "", fmt.Errorf("unknown template %q (choose app, library or diagnostic)", template)
	}
}
