package bodycodegen

import (
	"context"
	"fmt"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type SourceRenameEdit struct {
	Start  int    `json:"start_byte"`
	End    int    `json:"end_byte"`
	Kind   string `json:"kind"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type SourceRenameResult struct {
	Schema         string             `json:"schema"`
	Scope          string             `json:"scope"`
	ActivityID     string             `json:"activity_id"`
	From           string             `json:"from"`
	To             string             `json:"to"`
	IdentityPinned bool               `json:"identity_pinned"`
	InputSHA256    string             `json:"input_sha256"`
	OutputSHA256   string             `json:"output_sha256"`
	CallSites      int                `json:"call_sites"`
	Edits          []SourceRenameEdit `json:"edits"`
	Source         string             `json:"source"`
}

// RenameActivitySource produces a file-local edit proposal. It neither writes
// files nor executes bodies/models. Historical receipts retain their old hashes.
func RenameActivitySource(ctx context.Context, filename string, source []byte, from, to string) (SourceRenameResult, error) {
	var empty SourceRenameResult
	if ctx == nil {
		ctx = context.Background()
	}
	if len(source) > 128<<10 {
		return empty, fmt.Errorf("activity rename source exceeds 128 KiB")
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	tokens, diagnostics := syntax.Lex(to)
	if from == to || to == "_" || to == "init" || to == "let" || types.Universe.Lookup(to) != nil || !token.IsIdentifier(to) || diagnostics.HasErrors() ||
		len(tokens) != 2 || tokens[0].Kind != syntax.TokenIdentifier || tokens[0].Text != to {
		return empty, fmt.Errorf("choose a different Gooo activity identifier")
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() {
		return empty, diagnostics.Error()
	}
	if len(file.Imports) != 0 {
		return empty, fmt.Errorf("activity rename currently requires a single file without imports")
	}
	model, _, err := resolveBodyModel(file)
	if err != nil {
		return empty, err
	}
	var target *syntax.ActivityDecl
	for _, d := range file.Declarations {
		if a, ok := d.(*syntax.ActivityDecl); ok && a.Name == from {
			target = a
		}
	}
	if target == nil {
		return empty, fmt.Errorf("activity %q was not found", from)
	}
	result := SourceRenameResult{Schema: "gooo/activity-rename/v1", Scope: "single_source_file",
		From: from, To: to, InputSHA256: digest(source)}
	for _, n := range model.Nodes {
		if n.Name == to {
			return empty, fmt.Errorf("name %q already belongs to a declaration", to)
		}
		if n.Kind == bidir.ActivityKind && n.Name == from {
			result.ActivityID = string(n.ID)
		}
	}
	if result.ActivityID == "" {
		return empty, fmt.Errorf("activity has no resolved identity")
	}
	add := func(span syntax.Span, kind, replacement string) {
		start, end := span.Start.Offset, span.End.Offset
		result.Edits = append(result.Edits, SourceRenameEdit{start, end, kind, string(source[start:end]), replacement})
	}
	add(target.NameSpan, "declaration", to)
	if !target.IDPresent {
		span := target.Result.Span
		span.Start = span.End
		add(span, "pin_identity", " id "+renameString(result.ActivityID, ""))
		result.IdentityPinned = true
	}
	for _, d := range file.Declarations {
		if err := ctx.Err(); err != nil {
			return empty, err
		}
		a, ok := d.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if a.ValueProgramPresent {
			updated, count, err := renameBodyCalls(a.ValueProgram, len(a.Inputs), from, to)
			if err != nil {
				return empty, fmt.Errorf("activity %s: %w", a.Name, err)
			}
			result.CallSites += count
			if count != 0 {
				add(a.ValueProgramSpan, "body_calls", renameString(updated, string(source[a.ValueProgramSpan.Start.Offset:a.ValueProgramSpan.End.Offset])))
			}
		}
		if err := renameAssemblyReferences(a, source, from, to, &result, add); err != nil {
			return empty, err
		}
	}
	for _, binding := range file.Bindings {
		for _, endpoint := range []syntax.BindingEndpoint{binding.Producer, binding.Consumer} {
			if endpoint.PackageAlias != "" {
				return empty, fmt.Errorf("qualified bindings need a workspace rename")
			}
			if endpoint.Activity.Name == from {
				add(endpoint.Activity.Span, "binding", to)
			}
		}
	}
	slices.SortFunc(result.Edits, func(a, b SourceRenameEdit) int { return a.Start - b.Start })
	var out strings.Builder
	previous := 0
	for _, edit := range result.Edits {
		if edit.Start < previous {
			return empty, fmt.Errorf("overlapping rename edits")
		}
		out.Write(source[previous:edit.Start])
		out.WriteString(edit.After)
		previous = edit.End
	}
	out.Write(source[previous:])
	result.Source = out.String()
	afterFile, diags := ParseBodyFile(filename, []byte(result.Source))
	if diags.HasErrors() {
		return empty, fmt.Errorf("renamed source: %w", diags.Error())
	}
	after, _, err := resolveBodyModel(afterFile)
	if err != nil {
		return empty, fmt.Errorf("renamed source: %w", err)
	}
	if !renameIdentitiesPreserved(model, after, from, to) {
		return empty, fmt.Errorf("rename changed a semantic identity or PROV relation")
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	result.OutputSHA256 = digest([]byte(result.Source))
	return result, nil
}

func renameString(value, original string) string {
	if strings.HasPrefix(original, "`") && !strings.Contains(value, "`") {
		return "`" + value + "`"
	}
	var out strings.Builder
	out.WriteByte('"')
	for _, character := range value {
		switch character {
		case '\\', '"':
			out.WriteByte('\\')
			out.WriteRune(character)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if character < 0x20 {
				fmt.Fprintf(&out, `\u%04x`, character)
			} else {
				out.WriteRune(character)
			}
		}
	}
	out.WriteByte('"')
	return out.String()
}

func renameIdentitiesPreserved(before, after bidir.Model, from, to string) bool {
	if len(before.Nodes) != len(after.Nodes) || len(before.Relations) != len(after.Relations) || len(before.RuntimeBindings) != len(after.RuntimeBindings) {
		return false
	}
	for i, n := range before.Nodes {
		want := n.Name
		if n.Kind == bidir.ActivityKind && want == from {
			want = to
		}
		if after.Nodes[i].ID != n.ID || after.Nodes[i].Kind != n.Kind || after.Nodes[i].Name != want {
			return false
		}
	}
	relations := map[bidir.ID]bool{}
	for _, r := range before.Relations {
		relations[r.ID] = true
	}
	for _, r := range after.Relations {
		if !relations[r.ID] {
			return false
		}
	}
	for i, b := range before.RuntimeBindings {
		a := after.RuntimeBindings[i]
		if a.Feedback != b.Feedback || a.Entity != b.Entity || a.Producer.Activity.ID != b.Producer.Activity.ID ||
			a.Consumer.Activity.ID != b.Consumer.Activity.ID || a.Producer.Port.Name != b.Producer.Port.Name ||
			a.Consumer.Port.Name != b.Consumer.Port.Name {
			return false
		}
	}
	return true
}
