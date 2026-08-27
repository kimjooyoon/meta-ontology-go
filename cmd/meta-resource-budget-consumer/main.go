package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageresourcebudgetconsumer"
)

func main() {
	inputPath := flag.String("input", "", "raw producer evidence input")
	outputPath := flag.String("output", "", "independent consumer report output")
	sourceDir := flag.String("source-dir", "", "source directory for entry discovery")
	label := flag.String("label", "current-evidence", "evidence label")
	flag.Parse()
	if *sourceDir != "" && *inputPath == "" && *outputPath == "" {
		entry, err := languageresourcebudgetconsumer.DiscoverEntry(*sourceDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(entry)
		return
	}
	if *inputPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: meta-resource-budget-consumer -input INPUT -output OUTPUT [-label LABEL]")
		os.Exit(2)
	}
	input, err := languageresourcebudgetconsumer.ReadInput(*inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	report := languageresourcebudgetconsumer.Consume(input, *label)
	if err := languageresourcebudgetconsumer.WriteReport(*outputPath, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("independent consumer: label=%s decision=%s resolution=%s samples=%d/%d semantic=%s\n", report.Label, report.Decision, report.Resolution, report.Resource.Samples, report.Resource.ExpectedSamples, report.SemanticDecision)
	if report.Decision != "PASS" {
		os.Exit(1)
	}
}
