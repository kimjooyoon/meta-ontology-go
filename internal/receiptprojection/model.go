// Package receiptprojection compiles the bounded receipt schema profile from
// source-owned Gooo declarations. It does not activate EntityFields V1's public
// semantic profile, execute activities, call a model, or grant write authority.
package receiptprojection

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/token"
	"net/url"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const Profile = "gooo/receipt-structure-projection/v1"
const typePrefix = "gooo://receipt-structure/type/"

type Field struct {
	ID, Name, GoName, TypeID, GoType string
	Optional, Many                   bool
	Span                             syntax.Span
}
type Entity struct {
	ID, Name string
	Fields   []Field
}
type Projection struct {
	Package, Root, Schema, SourceSHA256 string
	Entities                            []Entity
	source                              []byte
	filename, originalRoot              string
}

var scalarTypes = map[string]string{
	typePrefix + "text": "string", typePrefix + "state": "string",
	typePrefix + "count": "int", typePrefix + "number": "float64",
	typePrefix + "record": "map[string]any", typePrefix + "counts": "map[string]int",
}

func Compile(filename string, source []byte, root string) (Projection, error) {
	if len(source) == 0 || len(source) > 64<<10 {
		return Projection{}, fmt.Errorf("receipt schema source must contain 1..65536 bytes")
	}
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, string(source), support)
	if file == nil || len(diagnostics) != 0 {
		return Projection{}, fmt.Errorf("receipt schema parse: %v", diagnostics)
	}
	if !exportedName(root) || file.Package.Name == "_" || !token.IsIdentifier(file.Package.Name) || token.Lookup(file.Package.Name).IsKeyword() {
		return Projection{}, fmt.Errorf("invalid Go package or root name")
	}
	if len(file.Bindings) != 0 {
		return Projection{}, fmt.Errorf("receipt schema profile does not execute bindings")
	}
	byName, ids, err := collectReceiptEntities(file)
	if err != nil {
		return Projection{}, err
	}
	schema, err := receiptSchema(byName, root)
	if err != nil {
		return Projection{}, err
	}
	entities, err := projectReceiptEntities(file, byName, ids, root)
	if err != nil {
		return Projection{}, err
	}
	return Projection{Package: file.Package.Name, Root: root, Schema: schema, Entities: entities,
		SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(source)), source: bytes.Clone(source),
		filename: filename, originalRoot: root}, nil
}

func exportedName(name string) bool {
	return token.IsIdentifier(name) && len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}
func goFieldName(name string) string {
	var b strings.Builder
	for part := range strings.SplitSeq(name, "_") {
		if part == "" {
			return ""
		}
		switch part {
		case "id", "sha", "rss", "cpu":
			b.WriteString(strings.ToUpper(part))
		default:
			b.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return b.String()
}
func claimID(ids map[string]bool, id string) error {
	u, err := url.Parse(id)
	if err != nil || u.Scheme == "" || strings.ContainsAny(id, " \t\r\n") || ids[id] {
		return fmt.Errorf("invalid or colliding stable ID %q", id)
	}
	ids[id] = true
	return nil
}
