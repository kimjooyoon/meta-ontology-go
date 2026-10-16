package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func observeExecution(cfg config) (*executionOrigin, error) {
	readGit := func(arguments ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, "git", append([]string{"--no-optional-locks", "-C", cfg.Root}, arguments...)...)
		command.WaitDelay = time.Second
		raw, err := command.Output()
		return strings.TrimSpace(string(raw)), err
	}
	checkout, headErr := readGit("rev-parse", "HEAD")
	status, statusErr := readGit("status", "--porcelain", "--untracked-files=all")
	value := executionOrigin{Kind: "LOCAL_OBSERVATION", CheckoutSHA: checkout,
		Clean: headErr == nil && statusErr == nil && status == ""}
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		if headErr != nil || statusErr != nil {
			return nil, fmt.Errorf("observe exact countercase checkout: %v / %v", headErr, statusErr)
		}
		env := map[string]string{}
		for _, key := range []string{"GITHUB_REPOSITORY", "GITHUB_RUN_ID", "GITHUB_RUN_ATTEMPT", "GITHUB_JOB",
			"GITHUB_WORKFLOW_SHA", "GITHUB_WORKFLOW_REF", "GITHUB_EVENT_NAME"} {
			env[key] = os.Getenv(key)
		}
		observed, err := ciExecution(cfg.Origin, env, checkout, value.Clean)
		if err != nil {
			return nil, err
		}
		value = observed
	}
	value.GoVersion, value.GOOS, value.GOARCH = runtime.Version(), runtime.GOOS, runtime.GOARCH
	value.GitObservationBudget = int64(30 * time.Second)
	return &value, nil
}
