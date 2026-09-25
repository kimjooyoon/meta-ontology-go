package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution/authorizationfoundation"
)

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	input := authorizationfoundation.Input{ExpectedSubject: options.subject}
	reads := []struct {
		path string
		out  *[]byte
	}{{options.foundation, &input.FoundationRaw}, {options.metadata, &input.MetadataRaw},
		{options.prior, &input.PriorReceiptRaw}, {options.current, &input.CurrentRaw}}
	for _, read := range reads {
		*read.out, err = os.ReadFile(read.path)
		if err != nil {
			return err
		}
	}
	receipt, err := authorizationfoundation.Evaluate(input)
	if err != nil {
		return err
	}
	suite := authorizationfoundation.EvaluateSuite(input)
	if suite.Decision != "PASS" || suite.Passed != suite.Total {
		return fmt.Errorf("authorization foundation suite failed closed")
	}
	receiptRaw, _ := json.MarshalIndent(receipt, "", "  ")
	suiteRaw, _ := json.MarshalIndent(suite, "", "  ")
	if err := writeOutside(options.receipt, append(receiptRaw, '\n')); err != nil {
		return err
	}
	if err := writeOutside(options.suite, append(suiteRaw, '\n')); err != nil {
		return err
	}
	return writeOutside(options.summary, authorizationfoundation.SummaryMarkdown(receipt, suite))
}
