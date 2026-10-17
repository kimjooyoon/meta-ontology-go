package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcerepack"
)

func main() {
	if err := sourcerepack.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "source-repacker:", err)
		os.Exit(1)
	}
}
