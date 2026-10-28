// record-origin-reader projects compiler observations and transports a source-
// owned query cursor. Gooo owns the work plan, edges, progress and result.
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

//go:embed main.gooo
var readerSource []byte

type progress struct {
	Status string `json:"status"`
	Steps  int    `json:"steps"`
	Budget int    `json:"step_budget"`
}

type queryResult struct {
	Schema              string            `json:"schema"`
	ReaderSourceSHA256  string            `json:"reader_source_sha256"`
	Steps               int               `json:"steps"`
	Budget              int               `json:"step_budget"`
	Result              json.RawMessage   `json:"result"`
	ActivityProjections map[string]string `json:"activity_projections"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	options := flag.NewFlagSet("record-origin-reader", flag.ContinueOnError)
	activity := options.String("activity", "", "source activity name")
	field := options.String("field", "", "stable output field ID")
	control := options.Bool("include-control", false, "follow guard and condition edges")
	history := options.Bool("include-history", false, "follow separate record version edges")
	if err := options.Parse(args); err != nil {
		return err
	}
	if options.NArg() != 1 || *activity == "" {
		return fmt.Errorf("usage: record-origin-reader --activity <name> [--field <stable-id>] [--include-control] [--include-history] <file.gooo>")
	}
	source, err := os.ReadFile(options.Arg(0))
	if err != nil {
		return err
	}
	report, err := bodycodegen.InspectTypedValueOrigins(options.Arg(0), source, *activity)
	if err != nil {
		return err
	}
	graph, err := project(report, *field, *control, *history)
	if err != nil {
		return err
	}
	result, err := query(ctx, readerSource, graph)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func query(ctx context.Context, source, graph []byte) (queryResult, error) {
	programs := make([]*bodycodegen.FixedProgram, 3)
	for i, name := range []string{"Begin", "Step", "Finish"} {
		program, err := bodycodegen.NewFixedProgram(ctx, "record-origin-reader.gooo", source, name)
		if err != nil {
			return queryResult{}, err
		}
		programs[i] = program
	}
	cursor, err := programs[0].Evaluate(ctx, tuple(graph))
	if err != nil {
		return queryResult{}, err
	}
	projections := map[string]string{cursor.ActivityID: cursor.ProjectionSHA256}
	state, err := cursorProgress(cursor.Value)
	if err != nil {
		return queryResult{}, err
	}
	for state.Status == "MORE" {
		if state.Budget < 1 || state.Steps < 0 || state.Steps >= state.Budget {
			return queryResult{}, fmt.Errorf("source cursor exhausted its declared work budget")
		}
		next, err := programs[1].Evaluate(ctx, tuple(graph, cursor.Value))
		if err != nil {
			return queryResult{}, err
		}
		observed, err := cursorProgress(next.Value)
		if err != nil {
			return queryResult{}, err
		}
		if observed.Steps != state.Steps+1 || observed.Budget != state.Budget {
			return queryResult{}, fmt.Errorf("source cursor did not preserve monotonic bounded progress")
		}
		projections[next.ActivityID] = next.ProjectionSHA256
		cursor, state = next, observed
	}
	result, err := programs[2].Evaluate(ctx, tuple(cursor.Value))
	if err != nil {
		return queryResult{}, err
	}
	projections[result.ActivityID] = result.ProjectionSHA256
	return queryResult{
		ActivityProjections: projections, Schema: "gooo/record-origin-query/v1",
		ReaderSourceSHA256: result.SourceSHA256, Steps: state.Steps,
		Budget: state.Budget, Result: result.Value,
	}, nil
}

func cursorProgress(raw []byte) (progress, error) {
	var result progress
	err := json.Unmarshal(raw, &result)
	return result, err
}

func tuple(values ...json.RawMessage) []byte {
	result, err := json.Marshal(values)
	if err != nil {
		panic(err)
	} // RawMessages originate from encoders or typed evaluation.
	return result
}
