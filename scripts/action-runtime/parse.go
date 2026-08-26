package main

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	usesPattern = regexp.MustCompile(
		`^\s*(?:-\s*)?uses:\s*([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)@([^\s#]+)`,
	)
	majorPattern = regexp.MustCompile(`^v([0-9]+)(?:\..*)?$`)
)

func parseWorkflow(source []byte) []useSite {
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	sites := make([]useSite, 0)
	for index, line := range lines {
		match := usesPattern.FindStringSubmatch(line)
		if match == nil || !strings.HasPrefix(match[1], "actions/") {
			continue
		}
		sites = append(sites, useSite{
			Action: match[1],
			Ref:    match[2],
			Line:   index + 1,
			Inputs: readInputs(lines, index+1, indentation(line)),
		})
	}
	sort.Slice(sites, func(i, j int) bool {
		if sites[i].Line != sites[j].Line {
			return sites[i].Line < sites[j].Line
		}
		return sites[i].Action < sites[j].Action
	})
	return sites
}

func parseMajor(reference string) (int, bool) {
	match := majorPattern.FindStringSubmatch(reference)
	if match == nil {
		return 0, false
	}
	major, err := strconv.Atoi(match[1])
	return major, err == nil
}
