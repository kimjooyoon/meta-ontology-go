package main

import (
	"encoding/json"
	"os"
	"strings"
)

func loadInputs(densityName, extractionName, observedName, untrackedName string) (densityReport, extractionReport, []string, []string, error) {
	var density densityReport
	if err := readJSON(densityName, &density); err != nil {
		return density, extractionReport{}, nil, nil, err
	}
	var extraction extractionReport
	if err := readJSON(extractionName, &extraction); err != nil {
		return density, extraction, nil, nil, err
	}
	observed, err := readPaths(observedName)
	if err != nil {
		return density, extraction, nil, nil, err
	}
	untracked, err := readPaths(untrackedName)
	return density, extraction, observed, untracked, err
}

func readPaths(name string) ([]string, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

func readJSON(name string, target any) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func writeOutputs(expectedName, reportName string, report evidence) error {
	expected := strings.Join(report.Expected, "\n")
	if expected != "" {
		expected += "\n"
	}
	if err := os.WriteFile(expectedName, []byte(expected), 0o644); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(reportName, append(data, '\n'), 0o644)
}
