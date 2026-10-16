// Command self-improvement-execution-consumer connects and consumes a bounded grant.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation"
)

type options struct {
	mode, output, input, reservation, repository, subject, workflow string
	run                                                             int64
	attempt                                                         int
}

func main() {
	var o options
	flag.StringVar(&o.mode, "mode", "prepare", "dispatch, prepare, reserve or consume")
	flag.StringVar(&o.output, "output", "", "caller-owned output directory (prepare) or JSON file")
	flag.StringVar(&o.input, "input", "", "prepared source or explicit source envelope")
	flag.StringVar(&o.reservation, "reservation", "", "reservation JSON")
	flag.StringVar(&o.repository, "repository", os.Getenv("GITHUB_REPOSITORY"), "owner/repository")
	flag.StringVar(&o.subject, "subject", os.Getenv("SUBJECT_SHA"), "logical subject")
	flag.StringVar(&o.workflow, "workflow", os.Getenv("GITHUB_SHA"), "producer workflow commit")
	flag.Int64Var(&o.run, "run", envInt("GITHUB_RUN_ID"), "current workflow run")
	flag.IntVar(&o.attempt, "attempt", envAttempt("GITHUB_RUN_ATTEMPT"), "current attempt")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	if err := run(ctx, o); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func envInt(key string) int64 { v, _ := strconv.ParseInt(os.Getenv(key), 10, 64); return v }

func envAttempt(key string) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value < 1 {
		return 0
	}
	return value
}

func run(ctx context.Context, o options) error {
	if o.output == "" {
		return fmt.Errorf("-output is required")
	}
	client := continuation.Client{Repository: o.repository, Token: os.Getenv("GH_TOKEN"), Root: os.DirFS(".")}
	switch o.mode {
	case "dispatch":
		return dispatch(ctx, client, o)
	case "prepare":
		return prepare(ctx, client, o)
	case "reserve":
		return reserve(o)
	case "consume":
		return consume(o)
	default:
		return fmt.Errorf("unknown consumer mode")
	}
}
