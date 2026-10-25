package sourcepartition

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type Size struct {
	Bytes int `json:"bytes"`
	Lines int `json:"physical_lines"`
}

type Boundary struct {
	Declaration syntax.Declaration
	Left, Right Size
}

type Layout struct {
	Input       Size
	HeaderBytes int
	Boundaries  []Boundary
}

// InspectBoundaries observes the exact byte partition without allocating every
// candidate's output. Sizes include copied headers, comments and blank lines.
// Declaration boundaries are syntactic candidates; consumer admission is separate.
func InspectBoundaries(filename, source string) (Layout, error) {
	file, header, err := partitionSyntax(filename, source)
	if err != nil {
		return Layout{}, err
	}
	totalNewlines := strings.Count(source, "\n")
	headerNewlines := strings.Count(source[:header], "\n")
	result := Layout{Input: physicalSize(source), HeaderBytes: header,
		Boundaries: make([]Boundary, 0, len(file.Decls)-1)}
	previous, prefixNewlines := 0, 0
	for _, declaration := range file.Decls[1:] {
		cut := declaration.SourceSpan().Start.Offset
		prefixNewlines += strings.Count(source[previous:cut], "\n")
		leftLines := prefixNewlines
		if source[cut-1] != '\n' {
			leftLines++
		}
		rightLines := headerNewlines + totalNewlines - prefixNewlines
		if source[len(source)-1] != '\n' {
			rightLines++
		}
		result.Boundaries = append(result.Boundaries, Boundary{Declaration: declaration,
			Left: Size{cut, leftLines}, Right: Size{header + len(source) - cut, rightLines}})
		previous = cut
	}
	return result, nil
}

func physicalSize(source string) Size {
	lines := strings.Count(source, "\n")
	if source != "" && !strings.HasSuffix(source, "\n") {
		lines++
	}
	return Size{Bytes: len(source), Lines: lines}
}
