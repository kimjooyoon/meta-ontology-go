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
	metrics := flag.String("metrics", "", "exact current line metrics")
	expected := flag.String("expected-sha", "", "required source SHA")
	output := flag.String("output", "", "split plan output")
	packageRecipe := flag.String("package-partition-recipe", "", "exact Go package partition recipe")
	flag.Parse()
	var err error
	if *packageRecipe != "" {
		err = runPackagePartition(*root, *packageRecipe, *expected, *output)
	} else if *metrics != "" {
		err = runMetrics(*root, *metrics, *expected, *output)
	} else {
		err = run(*root, *evidence, *expected, *output)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func runMetrics(root, metrics, expected, output string) error {
	if root == "" || metrics == "" || expected == "" || output == "" {
		return fmt.Errorf("root, metrics, expected-sha, and output are required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	inputs, err := loadMetricSubjects(metrics, expected)
	if err != nil {
		return err
	}
	plans := make([]planSubject, 0, len(inputs))
	counterexamples := make([]planCounterexample, 0)
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
			return fmt.Errorf("line evidence drift for %s: %d != %d", input.Logical, lines, input.Value)
		}
		atoms, parseErr := declarationAtoms(name, data)
		if parseErr != nil {
			return parseErr
		}
		plan, failures := classify(input, atoms)
		plans = append(plans, plan)
		counterexamples = append(counterexamples, failures...)
	}
	report := buildReport(expected, plans, counterexamples)
	if err := writeReport(output, report); err != nil {
		return err
	}
	return requireClassified(report)
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
	counterexamples := make([]planCounterexample, 0)
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
		plan, failures := classify(input, atoms)
		plans = append(plans, plan)
		counterexamples = append(counterexamples, failures...)
	}
	report := buildReport(expected, plans, counterexamples)
	if err := writeReport(output, report); err != nil {
		return err
	}
	counts := indicatorCounts(plans)
	density := counts["density-rewrite"] + counts["static-density-rewrite"] +
		counts["large-density-rewrite"]
	extraction := counts["no-movable-declaration"] + counts["fixed-declaration-capacity"] +
		counts["movable-declaration-capacity"]
	fmt.Printf("logical-split-planner: subjects=%d projectable=%d density=%d extraction=%d known_contradictions=%d\n",
		len(plans), counts["projectable"], density, extraction, counts["declaration-capacity-contradiction"])
	return requireClassified(report)
}
