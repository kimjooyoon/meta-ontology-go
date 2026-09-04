package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	contract := flag.String("contract", "", "continuity policy .gooo")
	project := flag.String("project", "", "utility project .gooo")
	manifest := flag.String("manifest", "", "utility evidence manifest")
	output := flag.String("output", "", "machine verification report")
	human := flag.String("human-output", "", "human verification dossier")
	flag.Parse()
	if *contract == "" || *project == "" || *manifest == "" || *output == "" || *human == "" {
		fail("usage: self-improvement-public-utility -contract FILE -project FILE -manifest FILE -output FILE -human-output FILE")
	}
	if err := verifyUtility(*contract, *project, *manifest, *output, *human); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
