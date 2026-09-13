package main

import (
	"context"
	"encoding/json"
	"errors"
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

func run(args []string, output, diagnostics io.Writer) int {
	flags := flag.NewFlagSet("goal-counterexample", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	policy := flags.String("policy", "", "original Gooo policy")
	goal := flags.String("goal", "", "separate pinned Gooo goal")
	request := flags.String("request", "", "explicit source and goal case snapshots")
	pkg := flags.String("profile-package", "metapolicycompilation", "expected package")
	namespace := flags.String("profile-namespace", "metapolicycompilation", "expected namespace")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *policy == "" || *goal == "" || *request == "" {
		fmt.Fprintln(diagnostics, "provide -policy, -goal and -request without positional arguments")
		return 2
	}
	var inputs [3][]byte
	for i, path := range []string{*policy, *goal, *request} {
		data, err := readInput(path)
		if err != nil {
			fmt.Fprintln(diagnostics, err)
			return 2
		}
		inputs[i] = data
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	report := policycompilation.ObserveGoalPolicyCounterexample(ctx, *policy, inputs[0], inputs[1], inputs[2], *pkg, *namespace)
	if err := json.NewEncoder(output).Encode(report); err != nil {
		fmt.Fprintln(diagnostics, err)
		return 2
	}
	switch report.State {
	case "PROPOSED", "NOT_PROPOSED":
		return 0
	case "REFUTED":
		return 1
	default:
		return 2
	}
}

func readInput(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 4<<20 {
		return nil, errors.New("input must be a regular file of at most 4 MiB")
	}
	data, err := io.ReadAll(io.LimitReader(file, (4<<20)+1))
	if err == nil && len(data) > 4<<20 {
		err = errors.New("input grew beyond 4 MiB")
	}
	return data, err
}
