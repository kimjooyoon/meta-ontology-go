package main

import (
	"flag"
	"fmt"
)

type options struct {
	root, head, readiness, concept, semantic, output, check string
}

func parseOptions(args []string) (options, error) {
	var value options
	set := flag.NewFlagSet("language-semantic-readiness-binding", flag.ContinueOnError)
	set.StringVar(&value.root, "root", "", "repository root")
	set.StringVar(&value.head, "head", "", "expected exact head SHA")
	set.StringVar(&value.readiness, "readiness", "", "language readiness artifact")
	set.StringVar(&value.concept, "concept", "", "language concept artifact")
	set.StringVar(&value.semantic, "semantic", "", "language semantic model artifact")
	set.StringVar(&value.output, "output", "", "output report outside repository")
	set.StringVar(&value.check, "check", "", "expected report to compare")
	if err := set.Parse(args); err != nil {
		return value, err
	}
	required := value.root != "" && value.head != "" && value.readiness != ""
	required = required && value.concept != "" && value.semantic != ""
	if !required || (value.output == "") == (value.check == "") {
		return value, fmt.Errorf("require root, head, three artifacts, and exactly one of output or check")
	}
	return value, nil
}
