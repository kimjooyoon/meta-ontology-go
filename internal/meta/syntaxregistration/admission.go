package syntaxregistration

import (
	"fmt"
	"go/ast"
	"strings"
)

const admissionChunkRows = 16

type admissionRow struct {
	version int
	key     string
	result  string
}

type admissionSet struct {
	root             ast.Expr
	prefix, argument string
	parameter, tag   string
	rows             []admissionRow
	helpers          []*ast.FuncDecl
}

func readAdmissionSets(source *goSource) ([]admissionSet, error) {
	digest := admissionSet{prefix: "acceptedDenominatorDigestChunk", argument: "digest",
		parameter: "string", tag: "digest"}
	header := admissionSet{prefix: "acceptedDenominatorHeaderChunk", argument: "value",
		parameter: "denominator", tag: "value.Version"}
	decode, err := source.function("decodeDenominator")
	if err != nil {
		return nil, err
	}
	validate, err := source.function("validateDenominator")
	if err != nil {
		return nil, err
	}
	if err := checkAdmissionRootBindings(source, decode, validate); err != nil {
		return nil, err
	}
	var digests []ast.Expr
	for _, statement := range decode.Body.List {
		if check, ok := statement.(*ast.IfStmt); ok && (strings.Contains(source.text(check.Cond), "digest != DenominatorDigest") ||
			strings.Contains(source.text(check.Cond), digest.prefix)) {
			if check.Init != nil {
				return nil, fmt.Errorf("denominator admission guard has a local initializer")
			}
			digests = append(digests, check.Cond)
		}
	}
	var headers []ast.Expr
	for _, statement := range validate.Body.List {
		if assignment, ok := statement.(*ast.AssignStmt); ok && len(assignment.Lhs) == 1 &&
			source.text(assignment.Lhs[0]) == "validHeader" && len(assignment.Rhs) == 1 {
			headers = append(headers, assignment.Rhs[0])
		}
	}
	if len(digests) != 1 || len(headers) != 1 {
		return nil, fmt.Errorf("denominator admission anchors are not exact")
	}
	digest.root, header.root = digests[0], headers[0]
	return readAdmissionPair(source, digest, header)
}

func readAdmissionPair(source *goSource, digest, header admissionSet) ([]admissionSet, error) {
	sets := []admissionSet{digest, header}
	for index := range sets {
		set := &sets[index]
		var err error
		if strings.Contains(source.text(set.root), set.prefix) {
			err = set.readChunks(source)
		} else {
			err = set.readLegacy(source)
		}
		if err != nil {
			return nil, err
		}
	}
	if len(sets[0].rows) != len(sets[1].rows) {
		return nil, fmt.Errorf("denominator admission inventories differ")
	}
	for index, row := range sets[0].rows {
		if row.version != sets[1].rows[index].version {
			return nil, fmt.Errorf("denominator admission version order differs")
		}
	}
	return sets, nil
}

func (set admissionSet) helperName(index int) string {
	return fmt.Sprintf("%s%d", set.prefix, index)
}

func (set admissionSet) rootCall() string {
	call := set.helperName(0) + "(" + set.argument + ")"
	if set.argument == "digest" {
		return "!" + call
	}
	return call
}
