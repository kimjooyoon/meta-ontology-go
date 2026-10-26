package main

import "strings"

type packageGenerateOptions struct {
	manifest, packageName, outputDir string
}

func parsePackageGenerateArgs(args []string) (options packageGenerateOptions, ok bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
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
