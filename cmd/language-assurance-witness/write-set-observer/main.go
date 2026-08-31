package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fail(fmt.Errorf("expected snapshot or compare command"))
	}
	var err error
	switch os.Args[1] {
	case "snapshot":
		err = snapshotCommand(os.Args[2:])
	case "compare":
		err = compareCommand(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fail(err)
	}
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
