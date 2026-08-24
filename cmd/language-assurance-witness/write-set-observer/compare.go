package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/writeset"
)

func compareCommand(arguments []string) error {
	flags := flag.NewFlagSet("compare", flag.ContinueOnError)
	beforePath := flags.String("before", "", "before snapshot")
	afterPath := flags.String("after", "", "after snapshot")
	declared := flags.String("declared", "", "comma-separated declared paths")
	output := flags.String("output", "", "receipt output path")
	expect := flags.String("expect", "PASS", "expected decision")
	subject := flags.String("subject-sha", "", "bound subject SHA")
	denominator := flags.String("denominator-digest", "", "bound denominator digest")
	if err := flags.Parse(arguments); err != nil { return err }
	if *beforePath == "" || *afterPath == "" || *output == "" { return fmt.Errorf("before, after, and output are required") }
	var before, after writeset.Snapshot
	if err := readJSON(*beforePath, &before); err != nil { return err }
	if err := readJSON(*afterPath, &after); err != nil { return err }
	paths := []string(nil)
	if strings.TrimSpace(*declared) != "" { paths = strings.Split(*declared, ",") }
	receipt := writeset.Compare(*subject, *denominator, before, after, paths)
	if err := writeJSON(*output, receipt); err != nil { return err }
	if receipt.Decision != *expect { return fmt.Errorf("decision %s, expected %s", receipt.Decision, *expect) }
	return nil
}
