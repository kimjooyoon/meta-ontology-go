package packageruntime

func initializationOrder(packages []PackageSpec) ([]string, error) {
	byPath := make(map[string]PackageSpec, len(packages))
	for _, spec := range packages {
		byPath[spec.Path] = spec
	}
	state := map[string]int{}
	order := make([]string, 0, len(packages))
	var visit func(string) error
	visit = func(packagePath string) error {
		if state[packagePath] == 1 {
			return reject("PACKAGE_IMPORT_CYCLE", "cycle at %q", packagePath)
		}
		if state[packagePath] == 2 {
			return nil
		}
		spec, exists := byPath[packagePath]
		if !exists {
			return reject("PACKAGE_IMPORT_UNKNOWN", "import %q", packagePath)
		}
		state[packagePath] = 1
		for _, imported := range spec.Imports {
			if err := visit(imported); err != nil {
				return err
			}
		}
		state[packagePath] = 2
		order = append(order, packagePath)
		return nil
	}
	for _, spec := range packages {
		if err := visit(spec.Path); err != nil {
			return nil, err
		}
	}
	return order, nil
}
