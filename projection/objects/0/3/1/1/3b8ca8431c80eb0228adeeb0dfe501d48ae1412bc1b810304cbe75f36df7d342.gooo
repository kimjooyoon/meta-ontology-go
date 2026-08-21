package main

import (
	"fmt"
	"os"
)

func run(opts options) error {
	if opts.root == "" {
		return fmt.Errorf("root must not be empty")
	}
	paths, err := collectTargets(opts)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stdout, "splitter: no failed source indicators")
		return nil
	}
	summary := 0
	for _, path := range paths {
		rewritten, err := refactor(path, opts)
		if err != nil {
			return err
		}
		summary += rewritten
	}
	if opts.write {
		fmt.Printf("splitter: rewritten=%d files\n", summary)
	} else {
		fmt.Printf("splitter: dry-run rewritten=%d files\n", summary)
	}
	return nil
}
