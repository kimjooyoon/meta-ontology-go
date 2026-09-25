package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorbinding"
)

func run(args []string, output io.Writer) error {
	value, err := parseConfig(args)
	if err != nil {
		return err
	}
	if err := validatePaths(value); err != nil {
		return err
	}
	before, err := readReport(value.before)
	if err != nil {
		return err
	}
	after, err := readReport(value.after)
	if err != nil {
		return err
	}
	if after.HeadSHA != value.expectedSHA {
		return fmt.Errorf("binding transition head mismatch")
	}
	transition, err := predecessorbinding.Compare(before, after)
	if err != nil {
		return err
	}
	replay, err := predecessorbinding.Compare(before, after)
	if err != nil || replay.Digest != transition.Digest {
		return fmt.Errorf("binding transition replay digest mismatch")
	}
	raw, err := encode(transition)
	if err != nil {
		return err
	}
	if err := writeOrCheck(value, raw); err != nil {
		return err
	}
	fmt.Fprintf(output, "predecessor-binding-transition: decision=%s static=%d->%d "+
		"dynamic=%d->%d bps=%d->%d unknown=%d writes=%d digest=%s\n",
		transition.Decision, transition.BeforeStatic, transition.AfterStatic,
		transition.BeforeDynamic, transition.AfterDynamic, transition.BeforeBPS,
		transition.AfterBPS, transition.Unknown, transition.RepositoryWrites,
		transition.Digest)
	return nil
}
