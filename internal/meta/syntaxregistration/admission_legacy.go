package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func admissionTerms(expression ast.Expr, operation token.Token) []ast.Expr {
	expression = ast.Unparen(expression)
	if binary, ok := expression.(*ast.BinaryExpr); ok && binary.Op == operation {
		return append(admissionTerms(binary.X, operation), admissionTerms(binary.Y, operation)...)
	}
	return []ast.Expr{expression}
}

func digestVersion(name string) (int, bool) {
	switch name {
	case "DenominatorDigest":
		return 21, true
	case "DenominatorMigrationDigest":
		return 22, true
	}
	text, found := strings.CutPrefix(name, "DenominatorMigrationV")
	text, suffix := strings.CutSuffix(text, "Digest")
	version, err := strconv.Atoi(text)
	return version, found && suffix && err == nil && version >= 23 && name == digestName(version)
}

func (set *admissionSet) readLegacy(source *goSource) error {
	operation := token.LOR
	if set.argument == "digest" {
		operation = token.LAND
	}
	for _, term := range admissionTerms(set.root, operation) {
		row, err := set.legacyRow(source, term)
		if err != nil {
			return err
		}
		if err := set.retainRow(row); err != nil {
			return err
		}
	}
	if admissionHelperCount(source, set.prefix) != 0 {
		return fmt.Errorf("denominator admission has orphan helpers")
	}
	return nil
}

func (set admissionSet) legacyRow(source *goSource, expression ast.Expr) (admissionRow, error) {
	if set.argument == "digest" {
		binary, ok := expression.(*ast.BinaryExpr)
		if ok && binary.Op == token.NEQ && source.text(binary.X) == "digest" {
			name, named := binary.Y.(*ast.Ident)
			if named {
				version, valid := digestVersion(name.Name)
				if valid {
					return admissionRow{version, name.Name, "true"}, nil
				}
			}
		}
	} else {
		terms := admissionTerms(expression, token.LAND)
		if len(terms) == 2 {
			identity, idOK := terms[0].(*ast.BinaryExpr)
			version, versionOK := terms[1].(*ast.BinaryExpr)
			if idOK && versionOK && identity.Op == token.EQL && version.Op == token.EQL &&
				source.text(identity.X) == "value.DenominatorID" && source.text(version.X) == "value.Version" {
				number, literal := integer(version.Y)
				id, err := strconv.Unquote(source.text(identity.Y))
				if literal && number >= 21 && err == nil && id == denominatorID(number) {
					return headerAdmissionRow(number), nil
				}
			}
		}
	}
	return admissionRow{}, fmt.Errorf("denominator admission has an unsupported predicate")
}

func headerAdmissionRow(version int) admissionRow {
	return admissionRow{version, strconv.Itoa(version), "value.DenominatorID == " + strconv.Quote(denominatorID(version))}
}

func (set *admissionSet) retainRow(row admissionRow) error {
	if len(set.rows) > 0 && set.rows[len(set.rows)-1].version >= row.version {
		return fmt.Errorf("denominator admission versions are repeated or unordered")
	}
	set.rows = append(set.rows, row)
	return nil
}
