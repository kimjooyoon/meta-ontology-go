package syntaxregistration

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strconv"
)

type boundary struct {
	ID            string `json:"id"`
	Schema        string `json:"schema"`
	MetaOperation string `json:"meta_operation"`
	Target        int    `json:"target"`
	LinkTarget    int    `json:"link_target"`
}

type denominator struct {
	Schema        string     `json:"schema"`
	DenominatorID string     `json:"denominator_id"`
	Version       int        `json:"version"`
	Boundaries    []boundary `json:"boundaries"`
}

func generateDenominator(raw []byte, record AppendRecord) ([]byte, error) {
	var value denominator
	if err := decodeStrict(raw, &value); err != nil {
		return nil, err
	}
	if value.Schema != "gooo/vertical-slice-boundary-denominator/v1" ||
		value.Version != record.BaseVersion || value.DenominatorID != denominatorID(record.BaseVersion) ||
		len(value.Boundaries) != 6 || value.Boundaries[0].ID != "syntax" ||
		record.CaseID == "" || record.Path == "" || record.SourceDigest == "" ||
		value.Boundaries[0].Target != record.CorpusTarget ||
		value.Boundaries[0].Target != record.BaseTarget || digest(raw) != record.BaselineDigest ||
		record.NextVersion != record.BaseVersion+1 || record.NextTarget != record.BaseTarget+1 ||
		!record.Accepted {
		return nil, fmt.Errorf("baseline denominator does not match the registered corpus")
	}
	links := 0
	for _, item := range value.Boundaries {
		links += item.LinkTarget
	}
	if links <= 0 {
		return nil, fmt.Errorf("baseline denominator has no link obligations")
	}
	baselineLinks := links
	value.Version = record.NextVersion
	value.DenominatorID = denominatorID(value.Version)
	value.Boundaries[0].Target = record.NextTarget
	links = 0
	for _, item := range value.Boundaries {
		links += item.LinkTarget
	}
	if links != baselineLinks {
		return nil, fmt.Errorf("generated denominator changed link obligations")
	}
	out, err := json.MarshalIndent(value, "", "  ")
	return append(out, '\n'), err
}

func denominatorID(version int) string {
	return fmt.Sprintf("gooo.denominator.capability.vertical-slice-closure.v%d", version)
}

func digestName(version int) string   { return fmt.Sprintf("DenominatorMigrationV%dDigest", version) }
func evidenceName(version int) string { return fmt.Sprintf("embeddedDenominatorV%d", version) }

func generateAdmission(source *goSource, version, capability int) error {
	trial := *source
	trial.edits = append([]sourceEdit(nil), source.edits...)
	sets, err := readAdmissionSets(&trial)
	if err != nil {
		return err
	}
	for _, set := range sets {
		if err := set.appendVersion(&trial, version); err != nil {
			return err
		}
	}
	if err := pinPreviousSyntaxTarget(&trial, version-1, capability); err != nil {
		return err
	}
	source.edits = trial.edits
	return nil
}

func generateDigest(source *goSource, previous []byte, version int, next []byte) error {
	bound := 0
	var anchor ast.Node
	for _, declaration := range source.file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || len(value.Values) != 1 {
				continue
			}
			if value.Names[0].Name == digestName(version) {
				return fmt.Errorf("new denominator digest is already declared")
			}
			if value.Names[0].Name == digestName(version-1) {
				literal, ok := value.Values[0].(*ast.BasicLit)
				if !ok {
					return fmt.Errorf("baseline digest is not literal")
				}
				text, err := strconv.Unquote(literal.Value)
				if err != nil || text != digest(previous) {
					return fmt.Errorf("baseline denominator digest mismatch")
				}
				anchor = value
				bound++
			}
		}
	}
	if bound != 1 {
		return fmt.Errorf("baseline denominator digest is not uniquely pinned")
	}
	source.appendAt(anchor, fmt.Sprintf("\nconst %s = %q\n", digestName(version), digest(next)))
	return nil
}
