package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	root := flag.String("root", "", "restored logical repository")
	evidence := flag.String("evidence", "", "exact projection evidence")
	expected := flag.String("expected-sha", "", "required source SHA")
	output := flag.String("output", "", "split plan output")
	flag.Parse()
	if err := run(*root, *evidence, *expected, *output); err != nil {
		log.Fatal(err)
	}
}

func run(root, evidence, expected, output string) error {
	if root == "" || evidence == "" || expected == "" || output == "" {
		return fmt.Errorf("root, evidence, expected-sha, and output are required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	inputs, err := loadSubjects(evidence, expected)
	if err != nil {
		return err
	}
	plans := make([]planSubject, 0, len(inputs))
	for _, input := range inputs {
		name, pathErr := sourcePath(absolute, input.Logical)
		if pathErr != nil {
			return pathErr
		}
		data, readErr := os.ReadFile(name)
		if readErr != nil {
			return readErr
		}
		if lines := physicalLines(data); lines != input.Value {
			return fmt.Errorf("line evidence drift for %s: %d != %d",
				input.Logical, lines, input.Value)
		}
		atoms, parseErr := declarationAtoms(name, data)
		if parseErr != nil {
			return parseErr
		}
		plans = append(plans, classify(input, atoms))
	}
	report := buildReport(expected, plans)
	if err := writeReport(output, report); err != nil {
		return err
	}
	counts := indicatorCounts(plans)
	density := counts["density-rewrite"] + counts["static-density-rewrite"] +
		counts["large-density-rewrite"]
	extraction := len(plans) - counts["projectable"] - density
	fmt.Printf("logical-split-planner: subjects=%d projectable=%d density=%d extraction=%d\n",
		len(plans), counts["projectable"], density, extraction)
	return requireClassified(report)
}
