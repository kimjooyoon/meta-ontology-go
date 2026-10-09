package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const bodyPlanUsage = "usage: gooo body-plan --source <source.gooo> [--entry <activity>] [--json | --inputs-template]"

func runBodyPlan(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	flags, mode, ok := parseBodyPlanArgs(args)
	if !ok {
		fmt.Fprintln(stderr, bodyPlanUsage)
		return exitUsage
	}
	fail := func(err error) int { fmt.Fprintf(stderr, "gooo body-plan: %v\n", err); return exitFailure }
	source, err := reader.ReadFile(flags["--source"])
	if err != nil {
		return fail(err)
	}
	plan, err := bodyexecution.InspectComposition(ctx, flags["--source"], source, flags["--entry"])
	if err != nil {
		return fail(err)
	}
	switch mode {
	case "--json":
		err = json.NewEncoder(stdout).Encode(plan)
	case "--inputs-template":
		var raw []byte
		raw, err = bodyexecution.CompositionInputTemplate(plan)
		if err == nil {
			_, err = fmt.Fprintln(stdout, string(raw))
		}
	default:
		err = printCompositionPlan(stdout, plan)
	}
	if err != nil {
		return fail(err)
	}
	return exitOK
}

func parseBodyPlanArgs(args []string) (map[string]string, string, bool) {
	flags, mode := map[string]string{"--source": "", "--entry": ""}, ""
	for i := 0; i < len(args); i++ {
		flag := args[i]
		if flag == "--json" || flag == "--inputs-template" {
			if mode != "" {
				return nil, "", false
			}
			mode = flag
			continue
		}
		value, known := flags[flag]
		if !known || value != "" || i+1 == len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			return nil, "", false
		}
		i++
		flags[flag] = args[i]
	}
	return flags, mode, flags["--source"] != ""
}

func printCompositionPlan(w io.Writer, plan bodyexecution.CompositionInspection) error {
	if _, err := fmt.Fprintf(w, "Source: %s\nCaller inputs:\n", plan.SourceSHA256); err != nil {
		return err
	}
	for _, input := range plan.CallerInputs {
		if _, err := fmt.Fprintf(w, "  %s: %s\n", input.Key, input.Type); err != nil {
			return err
		}
	}
	names := make(map[string]string, len(plan.Plan.Activities))
	if _, err := fmt.Fprintln(w, "Activity order:"); err != nil {
		return err
	}
	for _, activity := range plan.Plan.Activities {
		names[activity.ID] = activity.Name
		if _, err := fmt.Fprintf(w, "  %s -> %s\n", activity.Name, activity.OutputType); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "Bindings:"); err != nil {
		return err
	}
	for _, edge := range plan.Plan.Edges {
		if _, err := fmt.Fprintf(w, "  %s.%s -> %s.%s\n", names[edge.Producer], edge.ProducerPort,
			names[edge.Consumer], edge.ConsumerPort); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "Assembly order:"); err != nil {
		return err
	}
	for _, assembly := range plan.Assemblies {
		if _, err := fmt.Fprintf(w, "  %s: %s (%s; %d source cases, %d source holdout cases)\n",
			assembly.Name, assembly.Kind, assembly.Phase, assembly.SourceCases, assembly.SourceHoldoutCases); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "Inspection: 0 model calls, 0 candidate tests, 0 native executions.\n"+
		"Use --inputs-template for editable zero values. Body checks and correctness are separate.")
	return err
}
