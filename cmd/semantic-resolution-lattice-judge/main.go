package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	source := flag.String("source", "examples/semantic-resolution-lattice/main.gooo", "Gooo source")
	receipt := flag.String("receipt", "examples/semantic-resolution-lattice/receipt.json", "receipt")
	check := flag.Bool("check", false, "require a valid receipt")
	flag.Parse()
	if !*check {
		fatal(errors.New("-check is required for the independent adjudicator"))
	}
	if err := validate(*source, *receipt); err != nil {
		fatal(err)
	}
	fmt.Println("semantic resolution lattice: PASS")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
