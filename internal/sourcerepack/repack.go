package sourcerepack

import (
	"errors"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

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

type repackBlocked struct {
	Subject string
	Reason  string
}

func (blocked repackBlocked) Error() string { return blocked.Subject + ": " + blocked.Reason }
func (blocked repackBlocked) Unwrap() error { return errRepackBlocked }

func checkPlans(cfg config, report metricevidence.Report, indicators []metricevidence.Indicator) error {
	return checkPlansTo(os.Stdout, cfg, report, indicators)
}
