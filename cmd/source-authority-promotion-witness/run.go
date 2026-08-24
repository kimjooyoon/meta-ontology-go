package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthoritypromotion"
)

func run(args []string, stderr io.Writer) int {
	options, err := parseOptions(args)
	if err != nil { fmt.Fprintln(stderr, err); return 2 }
	assurance, err := readLimited(options.assurance)
	if err != nil { fmt.Fprintln(stderr, err); return 2 }
	upstream, err := readLimited(options.upstream)
	if err != nil { fmt.Fprintln(stderr, err); return 2 }
	report := sourceauthoritypromotion.Evaluate(sourceauthoritypromotion.Input{
		SubjectSHA: options.subjectSHA, AssuranceJSON: assurance, UpstreamJSON: upstream,
	})
	if err := writeExclusive(options.output, report); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	return 0
}
