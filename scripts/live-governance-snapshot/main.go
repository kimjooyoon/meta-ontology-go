package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/governancesnapshot"
)

type configuration struct {
	Snapshot string
	Root     string
	Contract string
	Graph    string
	Head     string
	Output   string
	Markdown string
	Check    bool
}

func main() {
	config := parseFlags()
	contract, err := readContract(config.Contract)
	if err != nil {
		die(err)
	}
	snapshot, err := governancesnapshot.LoadSnapshot(config.Snapshot, config.Root)
	if err != nil {
		die(err)
	}
	graph, err := readGraph(config.Graph)
	if err != nil {
		die(err)
	}
	report := governancesnapshot.Evaluate(snapshot, contract, graph)
	if err := writeReport(config, report); err != nil {
		die(err)
	}
	if config.Check {
		if err := governancesnapshot.ValidateReport(report, contract, graph, config.Head); err != nil {
			die(err)
		}
	}
}

func parseFlags() configuration {
	var config configuration
	flag.StringVar(&config.Snapshot, "snapshot", "", "captured public REST snapshot")
	flag.StringVar(&config.Root, "root", ".", "caller-owned snapshot root")
	flag.StringVar(&config.Contract, "contract", "examples/live-governance-snapshot/contract.json", "governance contract")
	flag.StringVar(&config.Graph, "graph", "", "materialized Gooo graph")
	flag.StringVar(&config.Head, "head", "", "expected checkout head")
	flag.StringVar(&config.Output, "output", "", "JSON report output")
	flag.StringVar(&config.Markdown, "markdown", "", "Markdown report output")
	flag.BoolVar(&config.Check, "check", false, "validate the report")
	flag.Parse()
	return config
}

func readContract(path string) (governancesnapshot.Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return governancesnapshot.Contract{}, err
	}
	var contract governancesnapshot.Contract
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return governancesnapshot.Contract{}, fmt.Errorf("decode contract: %w", err)
	}
	if err := governancesnapshot.ValidateContract(contract); err != nil {
		return governancesnapshot.Contract{}, err
	}
	return contract, nil
}

func readGraph(path string) (governancesnapshot.RawGraph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return governancesnapshot.RawGraph{}, err
	}
	var graph governancesnapshot.RawGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return governancesnapshot.RawGraph{}, fmt.Errorf("decode graph: %w", err)
	}
	return graph, nil
}

func writeReport(config configuration, report governancesnapshot.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if config.Output == "" {
		return fmt.Errorf("report output is required")
	}
	if err := os.MkdirAll(filepath.Dir(config.Output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(config.Output, data, 0o644); err != nil {
		return err
	}
	if config.Markdown != "" {
		if err := os.MkdirAll(filepath.Dir(config.Markdown), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(config.Markdown, []byte(report.HumanReport), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
