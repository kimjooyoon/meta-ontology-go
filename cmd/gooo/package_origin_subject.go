package main

func hasPackageOriginQueryOptions(flags map[string]string) bool {
	for _, name := range []string{"--origin-kind", "--origin-offset", "--origin-limit", "--origin-shape",
		"--origin-field-id", "--origin-activity-id", "--origin-root-activity-id"} {
		if flags[name] != "" {
			return true
		}
	}
	return false
}
