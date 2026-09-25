package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(parseConfig()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
