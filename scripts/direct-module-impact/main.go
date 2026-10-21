// Command direct-module-impact observes Gooo-owned direct dependency co-changes.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

func main() {
	root := flag.String("root", ".", "repository root")
	module := flag.String("module", "", "direct module path in payload.gooo")
	version := flag.String("version", "", "exact proposed version")
	destination := flag.String("materialize-dir", "", "write a complete candidate outside the repository")
	flag.Parse()
	if flag.NArg() != 0 || *module == "" {
		fmt.Fprintln(os.Stderr, "provide -module and -version")
		os.Exit(2)
	}
	var result any
	var err error
	if *destination == "" {
		result, err = toolchainrelease.ObserveDirectModuleImpact(*root, *module, *version)
	} else {
		result, err = toolchainrelease.ProjectDirectModuleCandidate(*root, *destination, *module, *version)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
