package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "source-repacker:", err)
		os.Exit(1)
	}
}

var errRepackBlocked = errors.New("repack blocked")

type fileEdit struct {
	Path    string
	Subject string
	Before  []byte
	After   []byte
	Mode    uint32
}

type repackPlan struct {
	Edits []fileEdit
}

type stagedEdit struct {
	temporary string
	edit      fileEdit
}

type importAddition struct {
	Name string
	Path string
}

func checkPlans(cfg config, report metricevidence.Report, indicators []metricevidence.Indicator) error {
	planned, blocked, matched := 0, 0, 0
	for _, indicator := range indicators {
		if cfg.subject != "" && cfg.subject != indicator.Subject {
			continue
		}
		matched++
		_, err := planRepack(cfg.root, indicator.Subject, report.Meta.Policy.MaxFileLines)
		if err == nil {
			planned++
			continue
		}
		if errors.Is(err, errRepackBlocked) {
			blocked++
			continue
		}
		return fmt.Errorf("plan %s: %w", indicator.Subject, err)
	}
	if cfg.subject != "" && matched == 0 {
		return fmt.Errorf("subject %q is not an actionable split indicator", cfg.subject)
	}
	fmt.Printf("source-repacker: actionable=%d planned=%d blocked=%d write=false\n", matched, planned, blocked)
	return nil
}
