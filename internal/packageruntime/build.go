package packageruntime

func Build(manifest Manifest) (Image, error) {
	image, _, err := build(manifest, false)
	return image, err
}

func build(manifest Manifest, includeInterface bool) (Image, []InterfacePackage, error) {
	normalized, err := normalizeManifest(manifest)
	if err != nil {
		return Image{}, nil, err
	}
	order, err := initializationOrder(normalized.Packages)
	if err != nil {
		return Image{}, nil, err
	}
	byPath := make(map[string]PackageSpec, len(normalized.Packages))
	for _, spec := range normalized.Packages {
		byPath[spec.Path] = spec
	}
	image := Image{Schema: ImageSchema, InitOrder: order}
	var interfaces []InterfacePackage
	activities := make([]EntryPlan, 0)
	exportsByPath := make(map[string][]Export, len(order))
	for _, packagePath := range order {
		compiled, compileErr := compilePackage(byPath[packagePath], exportsByPath, includeInterface)
		if compileErr != nil {
			return Image{}, nil, compileErr
		}
		if includeInterface {
			view, err := compiled.publicInterface()
			if err != nil {
				return Image{}, nil, err
			}
			interfaces = append(interfaces, view)
		}
		image.Packages = append(image.Packages, compiled.image)
		exportsByPath[packagePath] = compiled.image.Exports
		activities = append(activities, compiled.activities...)
	}
	entry, err := resolveEntry(normalized.Entry, byPath, activities)
	if err != nil {
		return Image{}, nil, err
	}
	image.Entry = entry
	image.Digest = imageDigest(image)
	return image, interfaces, nil
}

func resolveEntry(entry EntrySpec, packages map[string]PackageSpec, activities []EntryPlan) (EntryPlan, error) {
	if _, exists := packages[entry.PackagePath]; !exists {
		return EntryPlan{}, reject("ENTRY_PACKAGE_UNKNOWN", "package %q", entry.PackagePath)
	}
	for _, activity := range activities {
		if activity.PackagePath == entry.PackagePath && activity.Activity == entry.Activity {
			return activity, nil
		}
	}
	return EntryPlan{}, reject("ENTRY_ACTIVITY_UNKNOWN", "%s:%s", entry.PackagePath, entry.Activity)
}
