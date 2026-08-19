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
		return fmt.Errorf("no .go or .gooo files found")
	}
	summary := 0
	for _, path := range paths {
		rewritten, unsplittable, err := refactor(path, opts)
		if err != nil {
			return err
		}
		summary += rewritten
		if unsplittable > 0 {
			fmt.Fprintf(os.Stdout, "splitter: unsplittable declarations in %s: %d\n", path, unsplittable)
		}
	}
	if opts.write {
		fmt.Printf("splitter: rewritten=%d files\n", summary)
	} else {
		fmt.Printf("splitter: dry-run rewritten=%d files\n", summary)
	}
	return nil
}
