package main

import (
	"flag"
	"fmt"
)

type config struct {
	input   string
	report  string
	program string
	check   bool
}

func parse(args []string) (config, error) {
	flags := flag.NewFlagSet("integration-progress-witness", flag.ContinueOnError)
	var value config
	flags.StringVar(&value.input, "input", "", "raw observation JSON")
	flags.StringVar(&value.report, "report", "", "report JSON")
	flags.StringVar(&value.program, "program", "", "generated Gooo program")
	flags.BoolVar(&value.check, "check", false, "replay and compare existing outputs")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if value.input == "" || value.report == "" || value.program == "" || flags.NArg() != 0 {
		return config{}, fmt.Errorf("usage: integration-progress-witness -input observation.json -report report.json -program program.gooo [-check]")
	}
	return value, nil
}
