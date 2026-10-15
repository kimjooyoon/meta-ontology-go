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
	tokens := scanBodySurface(body)
	var edits bodySurfaceMap
	for i, current := range tokens {
		if edit, joined := elseSurfaceEdit(body, tokens, i); joined {
			edits = append(edits, edit)
		}
		if current.kind == token.IDENT && current.text == "let" {
			replacement := "var"
			if i > 0 && tokens[i-1].kind == token.FOR {
				replacement = ""
			}
			edits = append(edits, bodySurfaceEdit{current.start, current.end, replacement})
		}
		if current.kind == token.ASSIGN && i >= 3 && tokens[i-3].kind == token.FOR && tokens[i-2].text == "let" && tokens[i-1].kind == token.IDENT {
			edits = append(edits, bodySurfaceEdit{current.start, current.end, ":="})
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

type bodySurfaceToken struct {
	kind       token.Token
	text       string
	start, end int
}

func scanBodySurface(body string) []bodySurfaceToken {
	set := token.NewFileSet()
	file := set.AddFile("computes", -1, len(body))
	var scan scanner.Scanner
	scan.Init(file, []byte(body), nil, scanner.ScanComments)
	var tokens []bodySurfaceToken
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
		tokens = append(tokens, bodySurfaceToken{kind, text, start, start + len(text)})
	}
	return tokens
}

// NormalizedRange is used only to align a retained template fragment with its
// normalized counterpart. Expression coordinates are mapped back with OriginalRange.
func (m BodySourceMap) NormalizedRange(start, end int) FillOriginRange {
	return FillOriginRange{m.edits.normalizedOffset(start), m.edits.normalizedOffset(end)}
}

func (m bodySurfaceMap) normalizedOffset(original int) int {
	delta := 0
	for _, edit := range m {
		if original < edit.start {
			break
		}
		if original <= edit.end {
			if original == edit.end {
				return edit.start + delta + len(edit.text)
			}
			return edit.start + delta + min(original-edit.start, len(edit.text))
		}
		delta += len(edit.text) - (edit.end - edit.start)
	}
	return original + delta
}
