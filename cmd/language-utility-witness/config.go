package main

import (
	"flag"
	"fmt"
)

type config struct {
	contract, observation, report, program string
	check                                  bool
}

func parseConfig(args []string) (config, error) {
	set := flag.NewFlagSet("language-utility-witness", flag.ContinueOnError)
	var value config
	set.StringVar(&value.contract, "contract", "", "fixed language utility contract")
	set.StringVar(&value.observation, "observation", "", "CI evidence observation")
	set.StringVar(&value.report, "report", "", "language utility report")
	set.StringVar(&value.program, "program", "", "generated Gooo meta program")
	set.BoolVar(&value.check, "check", false, "compare with existing outputs")
	if err := set.Parse(args); err != nil {
		return config{}, err
	}
	if value.contract == "" || value.observation == "" || value.report == "" ||
		value.program == "" || len(set.Args()) != 0 {
		return config{}, fmt.Errorf("usage: language-utility-witness -contract contract.json " +
			"-observation observation.json -report report.json -program program.gooo [-check]")
	}
	return value, nil
}
