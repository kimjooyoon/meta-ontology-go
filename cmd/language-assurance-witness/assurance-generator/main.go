package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	specPath       = "internal/meta/languageassurance/write_set_metric.json"
	registryPath   = "internal/meta/languageassurance/registry.go"
	indicatorsPath = "internal/meta/languageassurance/indicators.go"
	generatedPath  = "internal/meta/languageassurance/write_set_generated.go"
)

func main() {
	write := flag.Bool("write", false, "write generated bindings")
	check := flag.Bool("check", false, "check generated bindings")
	flag.Parse()
	if *write == *check {
		fatalf("select exactly one of --write or --check")
	}
	spec, err := readMetricSpec(specPath)
	if err != nil {
		fatalf("read metric spec: %v", err)
	}
	registry, err := transformRegistry(registryPath, spec)
	if err != nil {
		fatalf("transform registry: %v", err)
	}
	indicators, err := transformIndicators(indicatorsPath)
	if err != nil {
		fatalf("transform indicators: %v", err)
	}
	generated, err := renderIndicator(spec)
	if err != nil {
		fatalf("render indicator: %v", err)
	}
	outputs := []outputFile{{registryPath, registry}, {indicatorsPath, indicators}, {generatedPath, generated}}
	if err := applyOutputs(outputs, *check); err != nil {
		fatalf("generated bindings: %v", err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
