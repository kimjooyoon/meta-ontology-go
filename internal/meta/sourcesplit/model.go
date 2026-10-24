package sourcesplit

import (
	"errors"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

var ErrBlocked = errors.New("split blocked")

// Blocker retains the planning cause without interpreting a process error string.
type Blocker struct {
	Reason string
	Detail string
}

func (b *Blocker) Error() string { return ErrBlocked.Error() + ": " + b.Detail }
func (b *Blocker) Unwrap() error { return ErrBlocked }

type Part struct {
	Path             string
	Subject          string
	Data             []byte
	DeclarationOrder []DeclarationOrder
}

type DeclarationOrder struct {
	Ordinal int    `json:"ordinal"`
	Digest  string `json:"digest"`
}

type Plan struct {
	Directory string
	Mode      os.FileMode
	Parts     []Part
}

func ValidateTopology(report metricevidence.Report, plan Plan) error {
	for _, directory := range report.Directories {
		if directory.Path != plan.Directory {
			continue
		}
		added := len(plan.Parts) - 1
		if directory.DirectFolders != 0 {
			return &Blocker{"MIXED_LOGICAL_TOPOLOGY", fmt.Sprintf("%s contains folders", plan.Directory)}
		}
		if directory.DirectFiles+added > report.Meta.Policy.MaxDirectEntries {
			return &Blocker{"LOGICAL_DIRECTORY_CAPACITY", fmt.Sprintf("%s projects %d entries", plan.Directory, directory.DirectFiles+added)}
		}
		return nil
	}
	return fmt.Errorf("metric evidence omits directory %q", plan.Directory)
}
