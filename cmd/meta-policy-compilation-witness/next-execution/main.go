// Command next-execution observes a source-pinned candidate in a fresh invocation.
// It is not a repository updater, historical attestor or revision admission gate.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("next-execution", flag.ContinueOnError)
	flags.SetOutput(stderr)
	policyPath := flags.String("policy", "", "candidate Gooo source, never generated Go")
	predecessorPath := flags.String("predecessor", "", "original independent-composition report")
	requestPath := flags.String("request", "", "pinned next-execution request and declared expectations")
	pkg := flags.String("profile-package", "metapolicycompilation", "expected Gooo package")
	namespace := flags.String("profile-namespace", "metapolicycompilation", "expected Gooo namespace")
	selection := bindGoalSelectionFlags(flags)
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *policyPath == "" || *predecessorPath == "" || *requestPath == "" {
		fmt.Fprintln(stderr, "-policy, -predecessor and -request are required; positional arguments are not accepted")
		return 2
	}
	if !selection.valid() {
		fmt.Fprintln(stderr, "-goal and -goal-digest are required together; materialization requires a goal")
		return 2
	}
	source, err := readBounded(*policyPath, 4<<20)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	predecessor, err := readBounded(*predecessorPath, 32<<20)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	request, err := readBounded(*requestPath, 4<<20)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if selection.enabled() {
		return selection.run(ctx, *policyPath, source, predecessor, request, *pkg, *namespace, stdout, stderr)
	}
	report := policycompilation.ObserveNextPolicyExecution(ctx, *policyPath, source, predecessor, request, *pkg, *namespace)
	if err := json.NewEncoder(stdout).Encode(report); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	switch report.Decision {
	case "NEXT_EXECUTION_OBSERVED":
		return 0
	case "REFUTED":
		return 1
	default:
		return 2
	}
}

func readBounded(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("%s must be a regular file no larger than %d bytes", path, limit)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err == nil && int64(len(data)) > limit {
		err = fmt.Errorf("%s exceeds the %d byte input bound", path, limit)
	}
	return data, err
}
