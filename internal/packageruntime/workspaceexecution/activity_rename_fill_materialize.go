package workspaceexecution

import (
	"fmt"
	"go/parser"
	"go/scanner"
	"go/token"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type renameFillPart struct {
	start, end, decodedStart int
	literal                  renameFillLiteral
	surface                  string
}

type renameFillHole struct {
	start, end int
	id         string
}

func materializeRenameFill(body string, bodySpan syntax.Span, candidate assemblyspec.FillCandidate,
	literals map[string]renameFillLiteral) (string, []renameFillPart, error) {
	fills := map[string]string{}
	for _, fill := range candidate.Fills {
		literal, ok := literals[fill.HoleID]
		if !ok || literal.value != fill.Expression {
			return "", nil, fmt.Errorf("fill %q has no original source literal", fill.HoleID)
		}
		if _, err := parser.ParseExpr(fill.Expression); err != nil {
			return "", nil, fmt.Errorf("fill %q expression: %w", fill.HoleID, err)
		}
		fills["__GOOO_BODY_HOLE_"+fill.HoleID+"__"] = fill.HoleID
	}
	holes, err := renameFillHolePositions(body, fills)
	if err != nil {
		return "", nil, err
	}
	var output strings.Builder
	var parts []renameFillPart
	previous := 0
	copyTemplate := func(end int) {
		if end == previous {
			return
		}
		start := output.Len()
		output.WriteString(body[previous:end])
		parts = append(parts, renameFillPart{start, output.Len(), previous, renameFillLiteral{bodySpan, body}, "computes"})
	}
	for _, hole := range holes {
		copyTemplate(hole.start)
		// Parentheses retain one expression and its original bytes. No Go
		// formatting or evaluation is used to invent a new source location.
		output.WriteByte('(')
		start := output.Len()
		literal := literals[hole.id]
		output.WriteString(literal.value)
		parts = append(parts, renameFillPart{start, output.Len(), 0, literal, "fill:" + candidate.ID + ":" + hole.id})
		output.WriteByte(')')
		previous = hole.end
	}
	copyTemplate(len(body))
	return output.String(), parts, nil
}

func renameFillHolePositions(body string, fills map[string]string) ([]renameFillHole, error) {
	file := token.NewFileSet().AddFile("computes", -1, len(body))
	var scan scanner.Scanner
	var scanError error
	scan.Init(file, []byte(body), func(position token.Position, message string) {
		if scanError == nil {
			scanError = fmt.Errorf("hole source at %d: %s", position.Offset, message)
		}
	}, scanner.ScanComments)
	counts := map[string]int{}
	var holes []renameFillHole
	for {
		pos, kind, literal := scan.Scan()
		if kind == token.EOF {
			break
		}
		id, ok := fills[literal]
		if kind != token.IDENT || !ok {
			continue
		}
		start := file.Offset(pos)
		holes = append(holes, renameFillHole{start, start + len(literal), id})
		counts[id]++
	}
	if scanError != nil {
		return nil, scanError
	}
	for _, id := range fills {
		if counts[id] != 1 {
			return nil, fmt.Errorf("expected one source_fill hole %q, found %d", id, counts[id])
		}
	}
	sort.Slice(holes, func(i, j int) bool { return holes[i].start < holes[j].start })
	return holes, nil
}
