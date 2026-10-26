package main

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
)

type packageGenerateOptions struct {
	manifest, packageName, outputDir string
	exports                          []bodylibrary.StageExport
	embedOrigin                      bool
}

func parsePackageGenerateArgs(args []string) (options packageGenerateOptions, ok bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--embed-origin":
			if options.embedOrigin {
				return options, false
			}
			options.embedOrigin = true
		case "--export-stage":
			if i+1 == len(args) {
				return options, false
			}
			i++
			name, id, found := strings.Cut(args[i], "=")
			if !found || name == "" || id == "" {
				return options, false
			}
			options.exports = append(options.exports, bodylibrary.StageExport{Name: name, ActivityID: id})
		case "--go-package", "--out":
			if i+1 == len(args) || strings.TrimSpace(args[i+1]) == "" || strings.HasPrefix(args[i+1], "-") {
				return options, false
			}
			i++
			target := &options.packageName
			if arg == "--out" {
				target = &options.outputDir
			}
			if *target != "" {
				return options, false
			}
			*target = args[i]
		default:
			if options.manifest != "" || strings.TrimSpace(arg) == "" || strings.HasPrefix(arg, "-") {
				return options, false
			}
			options.manifest = arg
		}
	}
	return options, options.manifest != "" && options.packageName != ""
}
