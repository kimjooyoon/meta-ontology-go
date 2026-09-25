package main

import (
	"flag"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/writeset"
)

func snapshotCommand(arguments []string) error {
	flags := flag.NewFlagSet("snapshot", flag.ContinueOnError)
	root := flags.String("root", "", "directory to observe")
	output := flags.String("output", "", "snapshot output path")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *root == "" || *output == "" {
		return fmt.Errorf("root and output are required")
	}
	snapshot, err := writeset.SnapshotDirectory(*root)
	if err != nil {
		return err
	}
	return writeJSON(*output, snapshot)
}
