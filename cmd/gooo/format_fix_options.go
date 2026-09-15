package main

const (
	formatUsage = "usage: gooo format [--check] [--json] <file.gooo>"
	fixUsage    = "usage: gooo fix [--json] <file.gooo>"
)

type formatOptions struct {
	filename string
	check    bool
	json     bool
}

func parseFormatOptions(args []string) (formatOptions, bool) {
	options := formatOptions{}
	for _, arg := range args {
		switch arg {
		case "--json":
			options.json = true
		case "--check":
			options.check = true
		default:
			if options.filename != "" || len(arg) == 0 || arg[0] == '-' {
				return formatOptions{}, false
			}
			options.filename = arg
		}
	}
	return options, options.filename != ""
}

func parseFixOptions(args []string) (formatOptions, bool) {
	options, ok := parseFormatOptions(args)
	return options, ok && !options.check
}
