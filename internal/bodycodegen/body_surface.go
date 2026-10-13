package bodycodegen

import (
	"fmt"
	"go/scanner"
	"go/token"
	"strings"
)

type bodySurfaceEdit struct {
	start, end int
	text       string
}

type bodySurfaceMap []bodySurfaceEdit

// The map stores token edits, not a second byte-sized copy of the source.
// Expressions and names retain their original byte ranges after lowering.
func rewriteBodySurface(body string) (string, bodySurfaceMap, error) {
	type lexeme struct {
		kind       token.Token
		text       string
		start, end int
	}
	set := token.NewFileSet()
	file := set.AddFile("computes", -1, len(body))
	var scan scanner.Scanner
	scan.Init(file, []byte(body), nil, scanner.ScanComments)
	var tokens []lexeme
	for {
		position, kind, text := scan.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.COMMENT {
			continue
		}
		if text == "" {
			text = kind.String()
		}
		start := file.Offset(position)
		tokens = append(tokens, lexeme{kind, text, start, start + len(text)})
	}
	var edits bodySurfaceMap
	for i, current := range tokens {
		if current.kind == token.IDENT && current.text == "let" {
			edits = append(edits, bodySurfaceEdit{current.start, current.end, "var"})
		}
		if current.kind == token.FOR && i+2 < len(tokens) && tokens[i+1].kind == token.IDENT &&
			tokens[i+2].kind == token.IDENT && tokens[i+2].text == "in" {
			edits = append(edits, bodySurfaceEdit{current.start, current.end, "for _,"})
		}
		if current.kind == token.IDENT && current.text == "in" && i >= 2 &&
			tokens[i-2].kind == token.FOR && tokens[i-1].kind == token.IDENT {
			edits = append(edits, bodySurfaceEdit{current.start, current.end, ":= range"})
		}
	}
	if len(edits) == 0 {
		return body, nil, nil
	}
	var out strings.Builder
	last := 0
	for _, edit := range edits {
		if edit.start < last || edit.end > len(body) {
			return "", nil, fmt.Errorf("overlapping body surface edits")
		}
		out.WriteString(body[last:edit.start])
		out.WriteString(edit.text)
		last = edit.end
	}
	out.WriteString(body[last:])
	return out.String(), edits, nil
}

func (m bodySurfaceMap) originalOffset(normalized int) int {
	delta := 0
	for _, edit := range m {
		start := edit.start + delta
		end := start + len(edit.text)
		if normalized < start {
			break
		}
		if normalized <= end {
			if normalized == end {
				return edit.end
			}
			return edit.start + min(normalized-start, edit.end-edit.start)
		}
		delta += len(edit.text) - (edit.end - edit.start)
	}
	return normalized - delta
}

func (m bodySurfaceMap) originalRange(span FillOriginRange) FillOriginRange {
	return FillOriginRange{Start: m.originalOffset(span.Start), End: m.originalOffset(span.End)}
}

// BodySourceMap translates parser offsets back to the decoded Gooo body.
// Token edits are private so callers cannot create a forged lowering map.
type BodySourceMap struct{ edits bodySurfaceMap }

func NormalizeBodySource(body string) (string, BodySourceMap, error) {
	text, edits, err := rewriteBodySurface(body)
	return text, BodySourceMap{edits}, err
}

func (m BodySourceMap) OriginalRange(start, end int) FillOriginRange {
	return m.edits.originalRange(FillOriginRange{start, end})
}
