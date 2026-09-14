package extractor

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strconv"
	"strings"
)

type mapLiteralCandidate struct {
	*suffixCandidate
	fusion *mapReceiverFusion
}

type mapLiteralEdit struct {
	start, end  int
	replacement string
}

type mapReceiverFusion struct {
	edit    mapLiteralEdit
	receipt mapReceiverFusionReceipt
}

type mapReceiverFusionReceipt struct {
	Rule                  string `json:"rule"`
	Binding               string `json:"binding"`
	SourceStart           int    `json:"source_start"`
	SourceEnd             int    `json:"source_end"`
	BindingUses           int    `json:"binding_uses"`
	SourceDigest          string `json:"source_digest"`
	ReplacementDigest     string `json:"replacement_digest"`
	BeforeRenderedOverage int    `json:"before_rendered_overage"`
	AfterRenderedOverage  int    `json:"after_rendered_overage"`
}

func renderMapLiteralCandidate(source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, literal *ast.CompositeLit, keys []string, evidence typeEvidence) (*mapLiteralCandidate, error) {
	start, end := fset.Position(literal.Pos()).Offset, fset.Position(literal.End()).Offset
	if start < 0 || end > len(source) || start >= end {
		return nil, returnTailContradiction(obligationControlFlow, "map literal source range is invalid")
	}
	name := mapLiteralHelperName(file, evidence.pkg, function.Name.Name)
	helper, call, err := renderMapLiteralParts(source, fset, literal, keys, name)
	if err != nil {
		return nil, err
	}
	edit := mapLiteralEdit{start: start, end: end, replacement: call}
	combined, err := renderMapLiteralProjection(source, edit, helper, nil)
	if err != nil {
		return nil, err
	}
	combined, fusion, err := prepareMapLiteralReceiver(source, combined, fset, file, function, evidence, edit, helper)
	if err != nil {
		return nil, err
	}
	return finishMapLiteralCandidate(source, combined, function.Name.Name, name, helper, edit, fusion)
}

func renderMapLiteralProjection(source []byte, literal mapLiteralEdit, helper []byte, fusion *mapReceiverFusion) ([]byte, error) {
	edits := []mapLiteralEdit{literal}
	if fusion != nil {
		edits = append(edits, fusion.edit)
		if edits[0].start > edits[1].start {
			edits[0], edits[1] = edits[1], edits[0]
		}
	}
	var output bytes.Buffer
	cursor := 0
	for _, edit := range edits {
		if edit.start < cursor || edit.end > len(source) || edit.start >= edit.end {
			return nil, returnTailContradiction(obligationControlFlow, "caller preparation source ranges overlap")
		}
		output.Write(source[cursor:edit.start])
		output.WriteString(edit.replacement)
		cursor = edit.end
	}
	output.Write(source[cursor:])
	output.WriteString("\n")
	output.Write(helper)
	combined, err := format.Source(output.Bytes())
	if err != nil {
		return nil, returnTailContradiction(obligationProjectedConform, "map construction candidate did not render")
	}
	return combined, nil
}

func prepareMapLiteralReceiver(source, combined []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, evidence typeEvidence, edit mapLiteralEdit, helper []byte) ([]byte, *mapReceiverFusion, error) {
	outer, err := mapLiteralMeasurement(combined, function.Name.Name)
	if err != nil || outer.overage == 0 {
		return combined, nil, err
	}
	fusion := findMapReceiverFusion(source, fset, file, function, evidence)
	if fusion == nil {
		return combined, nil, nil
	}
	prepared, err := renderMapLiteralProjection(source, edit, helper, fusion)
	if err != nil {
		return nil, nil, err
	}
	after, err := mapLiteralMeasurement(prepared, function.Name.Name)
	if err != nil {
		return nil, nil, err
	}
	if after.overage >= outer.overage {
		return combined, nil, nil
	}
	fusion.receipt.BeforeRenderedOverage = outer.overage
	fusion.receipt.AfterRenderedOverage = after.overage
	return prepared, fusion, nil
}

func finishMapLiteralCandidate(source, combined []byte, function, name string, helper []byte, edit mapLiteralEdit, fusion *mapReceiverFusion) (*mapLiteralCandidate, error) {
	before, err := mapLiteralMeasurement(source, function)
	if err != nil {
		return nil, err
	}
	outer, err := mapLiteralMeasurement(combined, function)
	if err != nil {
		return nil, err
	}
	inner, err := mapLiteralMeasurement(combined, name)
	if err != nil {
		return nil, err
	}
	after := outer.overage + inner.overage
	if !renderedCapacityProgress(renderedCapacitySnapshot{overage: before.overage}, renderedCapacitySnapshot{overage: after}) {
		return nil, returnTailContradiction(obligationRenderedCapacity, "map construction does not strictly reduce total rendered overage")
	}
	return &mapLiteralCandidate{suffixCandidate: &suffixCandidate{
		start: edit.start, end: edit.end, helperName: name, helper: helper, result: combined,
		beforeRenderedCapacityOverage: before.overage, afterRenderedCapacityOverage: after,
		renderedHelper: inner, renderedOuter: outer,
	}, fusion: fusion}, nil
}

func renderMapReceiverFusion(source []byte, fset *token.FileSet, first, next *ast.AssignStmt, receiver *ast.Ident) *mapReceiverFusion {
	start, end := fset.Position(first.Pos()).Offset, fset.Position(next.End()).Offset
	initStart, initEnd := fset.Position(first.Rhs[0].Pos()).Offset, fset.Position(first.Rhs[0].End()).Offset
	nextStart := fset.Position(next.Pos()).Offset
	receiverStart, receiverEnd := fset.Position(receiver.Pos()).Offset, fset.Position(receiver.End()).Offset
	if start < 0 || initStart < start || initEnd > nextStart || nextStart > receiverStart ||
		receiverStart >= receiverEnd || receiverEnd > end || end > len(source) {
		return nil
	}
	replacement := string(source[nextStart:receiverStart]) + string(source[initStart:initEnd]) + string(source[receiverEnd:end])
	return &mapReceiverFusion{
		edit: mapLiteralEdit{start: start, end: end, replacement: replacement},
		receipt: mapReceiverFusionReceipt{
			Rule: "caller-local-adjacent-pointer-receiver", Binding: receiver.Name,
			SourceStart: start, SourceEnd: end, BindingUses: 1,
			SourceDigest:      fmt.Sprintf("sha256:%x", sha256.Sum256(source[start:end])),
			ReplacementDigest: fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(replacement))),
		},
	}
}

func renderMapLiteralParts(source []byte, fset *token.FileSet, literal *ast.CompositeLit, keys []string, name string) ([]byte, string, error) {
	parameters := make([]string, 0, len(keys))
	arguments := make([]string, 0, len(keys))
	var body strings.Builder
	for index, key := range keys {
		pair := literal.Elts[index].(*ast.KeyValueExpr)
		start, end := fset.Position(pair.Value.Pos()).Offset, fset.Position(pair.Value.End()).Offset
		if start < 0 || end > len(source) || start >= end {
			return nil, "", returnTailContradiction(obligationFreeBindings, "map value source range is invalid")
		}
		parameter := fmt.Sprintf("value%d", index)
		parameters = append(parameters, parameter+" any")
		arguments = append(arguments, string(source[start:end]))
		fmt.Fprintf(&body, "\t\t%s: %s,\n", strconv.Quote(key), parameter)
	}
	helper := "func " + name + "(" + strings.Join(parameters, ", ") + ") map[string]any {\n\treturn map[string]any{\n" + body.String() + "\t}\n}\n"
	formatted, err := format.Source([]byte(helper))
	if err != nil {
		return nil, "", returnTailContradiction(obligationProjectedConform, "map constructor did not render")
	}
	return formatted, name + "(" + strings.Join(arguments, ", ") + ")", nil
}

func mapLiteralMeasurement(source []byte, name string) (renderedCapacityMeasurement, error) {
	rendered, err := renderedFunctionHelper(source, name)
	if err != nil {
		return renderedCapacityMeasurement{}, err
	}
	return canonicalRenderedCapacity(rendered)
}
