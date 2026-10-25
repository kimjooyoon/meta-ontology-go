// Package embeduse observes Go embed consumers of Gooo sources through the Go
// tool's selected package. Each variable is queried separately: a package-wide
// union can hide a single-file reader behind a sibling wildcard reader.
package embeduse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

type File struct {
	sourcebundle.Member
	Projected bool
}

type Consumer struct {
	Name, Kind     string
	DeclaredType   string
	Declaration    bidir.SourceSpan
	GoSource       int
	Members        []int
	AllMemberCount int
}

// File indices share one bounded table; names and locations identify Go source
// occurrences, not Gooo semantic IDs or runtime read permissions.
type Report struct {
	Projection                                []int
	RequestedDirectory, Directory, ImportPath string
	Build                                     map[string]string
	Files                                     []File
	Consumers                                 []Consumer
	IgnoredGoFiles                            []string
}

// Inspect observes one Go package in the current build context. Projected Gooo
// members are visible only to Go's overlay, never installed in the source tree.
// Other build contexts, dynamic reads and runtime interpretation remain outside
// this observation. Neither a complete repository inventory nor rewrite approval
// is implied. An error returns no report.
func Inspect(ctx context.Context, directory string, projected []sourcebundle.Source) (Report, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	directory, err := filepath.Abs(directory)
	if err != nil {
		return Report{}, err
	}
	root, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return Report{}, err
	}
	work, err := os.MkdirTemp("", "gooo-embed-use-")
	if err != nil {
		return Report{}, err
	}
	defer os.RemoveAll(work)
	observer := observer{ctx: ctx, root: root, work: work, env: os.Environ(), projections: map[string]string{}, inputs: map[string]int{}}
	if err := observer.prepareProjection(projected); err != nil {
		return Report{}, err
	}
	return observer.inspect(directory)
}

func (o *observer) inspect(requested string) (Report, error) {
	var build map[string]string
	if err := o.query(&build, "env", "-json", "GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOFLAGS", "GOEXPERIMENT"); err != nil {
		return Report{}, err
	}
	if strings.Contains(build["GOFLAGS"], "-overlay") {
		return Report{}, fmt.Errorf("Go embed observation cannot compose an overlay supplied through GOFLAGS")
	}
	selected, err := o.list(o.projections)
	if err != nil {
		return Report{}, err
	}
	o.report = Report{RequestedDirectory: requested, Directory: o.root, ImportPath: selected.ImportPath,
		Build: build, IgnoredGoFiles: append([]string{}, selected.IgnoredGoFiles...)}
	for _, name := range o.projectedNames {
		_, index, err := o.read(name)
		if err != nil {
			return Report{}, err
		}
		o.report.Projection = append(o.report.Projection, index)
	}
	declarations, masked, err := o.declarations(selected)
	if err != nil {
		return Report{}, err
	}
	for _, declaration := range declarations {
		if err := o.observeDeclaration(selected, declaration, masked); err != nil {
			return Report{}, err
		}
	}
	current, err := o.list(o.projections)
	if err != nil {
		return Report{}, err
	}
	if !reflect.DeepEqual(selected, current) {
		return Report{}, fmt.Errorf("Go package selection changed during observation")
	}
	if err := o.recheckInputs(); err != nil {
		return Report{}, err
	}
	return o.report, nil
}
