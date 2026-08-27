package main

import (
	"flag"
	"log"
)

type options struct {
	receiptPath     string
	archivePath     string
	verification    string
	verifierExit    int
	verifierVersion string
	output          string
	check           bool
}

func main() {
	var options options
	flag.StringVar(&options.receiptPath, "transport-receipt", "", "prior EHT-8 transport receipt")
	flag.StringVar(&options.archivePath, "source-archive", "", "downloaded producer artifact archive")
	flag.StringVar(&options.verification, "verification", "", "gh attestation verification JSON")
	flag.IntVar(&options.verifierExit, "verifier-exit-code", 0, "gh attestation verifier exit code")
	flag.StringVar(&options.verifierVersion, "verifier-version", "", "gh verifier version")
	flag.StringVar(&options.output, "output", "", "resolution receipt output")
	flag.BoolVar(&options.check, "check", false, "validate the emitted receipt")
	flag.Parse()
	if err := run(options); err != nil {
		log.Fatal(err)
	}
}
