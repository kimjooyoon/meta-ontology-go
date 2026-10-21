package replay

import (
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func bindContext(repository fs.FS, path string, sources []string, file, replayed *syntax.File, support syntax.EntityFieldsSupport) (*syntax.File, *syntax.File, error) {
	if file.Package == nil || file.Namespace == nil {
		return nil, nil, fmt.Errorf("primary source has no package headers")
	}
	declarations := make([]syntax.Declaration, 0)
	bindings := make([]syntax.BindingDecl, 0)
	imports := make([]syntax.ImportDecl, 0)
	seen := map[string]bool{path: true}
	for _, source := range sources {
		if source == "" || seen[source] {
			return nil, nil, fmt.Errorf("invalid companion source %q", source)
		}
		seen[source] = true
		raw, err := fs.ReadFile(repository, source)
		if err != nil {
			return nil, nil, fmt.Errorf("read companion %q: %w", source, err)
		}
		companion, diagnostics := syntax.ParseFileWithEntityFieldsSupport(source, string(raw), support)
		if diagnostics.HasErrors() {
			return nil, nil, fmt.Errorf("parse companion %q: %s", source, diagnostics.Error())
		}
		if companion.Package == nil || companion.Namespace == nil || companion.Package.Name != file.Package.Name || companion.Namespace.Name != file.Namespace.Name {
			return nil, nil, fmt.Errorf("companion %q has different package headers", source)
		}
		declarations = append(declarations, companion.Decls...)
		bindings = append(bindings, companion.Bindings...)
		imports = append(imports, companion.Imports...)
	}
	join := func(primary *syntax.File) *syntax.File {
		combined := primary.Clone()
		combined.Decls = append(append([]syntax.Declaration(nil), declarations...), combined.Decls...)
		combined.Declarations = combined.Decls
		combined.Bindings = append(append([]syntax.BindingDecl(nil), bindings...), combined.Bindings...)
		combined.Imports = append(append([]syntax.ImportDecl(nil), imports...), combined.Imports...)
		return combined
	}
	return join(file), join(replayed), nil
}
