package packageruntime

type compiledPackage struct {
	image      PackageImage
	activities []EntryPlan
}

func compilePackage(spec PackageSpec) (compiledPackage, error) {
	compiled := compiledPackage{image: PackageImage{
		Path: spec.Path, Name: spec.Name, Imports: append([]string(nil), spec.Imports...),
	}}
	declarations := map[string]bool{}
	for _, source := range spec.Sources {
		image, namespace, names, activities, err := compileSource(spec, source)
		if err != nil {
			return compiledPackage{}, err
		}
		if compiled.image.Namespace != "" && compiled.image.Namespace != namespace {
			return compiledPackage{}, reject("PACKAGE_NAMESPACE_MISMATCH", "package %q", spec.Path)
		}
		compiled.image.Namespace = namespace
		for _, name := range names {
			if declarations[name] {
				return compiledPackage{}, reject("PACKAGE_DECLARATION_DUPLICATE", "%s:%s", spec.Path, name)
			}
			declarations[name] = true
		}
		compiled.image.Sources = append(compiled.image.Sources, image)
		compiled.image.Declarations += image.Declarations
		compiled.activities = append(compiled.activities, activities...)
	}
	compiled.image.SemanticDigest = digestValue(struct {
		Path, Namespace string
		Sources         []SourceImage
	}{spec.Path, compiled.image.Namespace, compiled.image.Sources})
	return compiled, nil
}
