package packageruntime

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

// Same-file bindings retain their existing BX validation. Edges involving
// another source file are validated by the package export catalog and retain
// original source spans and activity IDs in PackageImage.Bindings.
func sameFileBindings(bindings []syntax.BindingDecl, file *syntax.File) []syntax.BindingDecl {
	names := map[string]bool{}
	for _, declaration := range file.Declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok {
			names[activity.Name] = true
		}
	}
	var result []syntax.BindingDecl
	for _, binding := range localBindings(bindings) {
		if names[binding.Producer.Activity.Name] && names[binding.Consumer.Activity.Name] {
			result = append(result, binding)
		}
	}
	return result
}
