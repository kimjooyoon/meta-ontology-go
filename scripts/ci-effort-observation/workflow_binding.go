package main

import (
	"fmt"
	"strings"
)

func bindWorkflowCommand(source []byte, path string, spec OperationSpec) (string, error) {
	job, ok := namedYAMLBlock(string(source), "name: "+spec.JobName, 2)
	if !ok {
		return "", fmt.Errorf("workflow job binding is missing for %s", spec.ID)
	}
	step, ok := namedYAMLBlock(job, "- name: "+spec.StepName, 6)
	run := ""
	if ok {
		run, ok = workflowRunText(step)
	} else {
		run, ok = unnamedWorkflowRun(job, spec.Command)
	}
	if !ok || !commandMatches(run, spec.Command) {
		return "", fmt.Errorf("workflow command binding is missing for %s", spec.ID)
	}
	evidence := struct {
		Path, SourceDigest, Job, Step, Run string
		Command                            []string
	}{path, digestBytes(source), spec.JobName, spec.StepName, run, spec.Command}
	return digestJSON(evidence), nil
}

func unnamedWorkflowRun(job string, command []string) (string, bool) {
	for line := range strings.SplitSeq(job, "\n") {
		value := strings.TrimSpace(line)
		if indentation(line) != 6 || !strings.HasPrefix(value, "- run:") {
			continue
		}
		run := strings.TrimSpace(strings.TrimPrefix(value, "- run:"))
		if commandMatches(run, command) {
			return run, true
		}
	}
	return "", false
}

func namedYAMLBlock(source, marker string, indent int) (string, bool) {
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		header := strings.TrimSpace(line)
		if indentation(line) != indent || (header != marker && !strings.HasSuffix(header, ":")) {
			continue
		}
		end := len(lines)
		for candidate := index + 1; candidate < len(lines); candidate++ {
			if strings.TrimSpace(lines[candidate]) != "" && indentation(lines[candidate]) <= indent {
				end = candidate
				break
			}
		}
		for candidate := index + 1; candidate < end; candidate++ {
			if strings.TrimSpace(lines[candidate]) == marker {
				return strings.Join(lines[index:end], "\n"), true
			}
		}
	}
	return "", false
}

func indentation(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func workflowRunText(step string) (string, bool) {
	lines := strings.Split(step, "\n")
	for index, line := range lines {
		if indentation(line) != 8 || !strings.HasPrefix(strings.TrimSpace(line), "run:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "run:"))
		if value != "|" {
			return value, value != ""
		}
		var body []string
		for _, candidate := range lines[index+1:] {
			if strings.TrimSpace(candidate) != "" && indentation(candidate) <= 8 {
				break
			}
			body = append(body, strings.TrimSpace(candidate))
		}
		return strings.Join(body, "\n"), len(body) > 0
	}
	return "", false
}

func commandMatches(run string, expected []string) bool {
	want := append([]string(nil), expected...)
	for line := range strings.SplitSeq(run, "\n") {
		got := commandTokens(line)
		if len(got) < len(want) {
			continue
		}
		match := true
		for index := range want {
			if got[index] != want[index] {
				match = false
				break
			}
		}
		if match && (len(got) == len(want) || commandDelimiter(got[len(want)])) {
			return true
		}
	}
	return false
}

func commandTokens(value string) []string {
	value = strings.ReplaceAll(value, "\"", "")
	value = strings.ReplaceAll(value, "'", "")
	return strings.Fields(value)
}

func commandDelimiter(value string) bool {
	return value == "|" || value == "||" || value == "&&" || value == ";"
}
