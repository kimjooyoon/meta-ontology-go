package main

import (
	"fmt"
	"os"

	capability "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution"
)

func run(arguments []string) error {
	options, err := parseOptions(arguments)
	if err != nil {
		return err
	}
	var observation capability.Observation
	if options.Mode == "observe" {
		parent, err := os.ReadFile(options.ParentReport)
		if err != nil {
			return err
		}
		observation = capability.Observe(capability.ObserverOptions{
			SubjectSHA: options.SubjectSHA, SourceRoot: options.SourceRoot,
			ExternalRoot: options.ExternalRoot,
		}, parent)
		if err := writeJSON(options.Observation, observation); err != nil {
			return err
		}
	} else {
		observation = capability.LoadObservation(options.Observation)
	}
	report := capability.Evaluate(observation)
	suite := capability.RunSuite(observation.SubjectSHA)
	if err := writeJSON(options.Report, report); err != nil {
		return err
	}
	if err := writeJSON(options.Suite, suite); err != nil {
		return err
	}
	if report.Decision != capability.DecisionExecutable || suite.Passed != suite.Total {
		return fmt.Errorf("capability conformance failed: %s/%s %d/%d",
			report.Decision, report.Resolution, suite.Passed, suite.Total)
	}
	return nil
}
