package main

func parseQueryStringArgument(options queryOptions, arg, value string) (queryOptions, string) {
	var target *string
	switch arg {
	case "--root":
		target = &options.root
	case "--target":
		target = &options.target
	case "--relation":
		target = &options.relation
	case "--rule":
		target = &options.rule
	case "--layer":
		target = &options.layer
	case "--direction":
		target = &options.direction
	default:
		options.operation = "invalid"
		return options, ""
	}
	if *target != "" {
		options.operation = "invalid"
		return options, ""
	}
	*target = value
	return options, ""
}
func parseQueryBoundArgument(options queryOptions, arg, value string) (queryOptions, string) {
	if arg == "--limit" {
		if options.limitSet {
			options.limit = 0
			return options, ""
		}
		limit, err := parseQueryInteger(value)
		if err != nil {
			options.limit, options.limitSet = 0, true
			return options, ""
		}
		options.limit, options.limitSet = limit, true
		return options, ""
	}
	if options.maxDepthSet {
		options.maxDepth = 0
		return options, ""
	}
	depth, err := parseQueryInteger(value)
	if err != nil {
		options.maxDepth, options.maxDepthSet = 0, true
		return options, ""
	}
	options.maxDepth, options.maxDepthSet = depth, true
	return options, ""
}
