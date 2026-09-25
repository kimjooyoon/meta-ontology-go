package packageruntime

func Build(manifest Manifest) (Image, error) {
	normalized, err := normalizeManifest(manifest)
	if err != nil {
		return Image{}, err
	}
	order, err := initializationOrder(normalized.Packages)
	if err != nil {
		return Image{}, err
	}
	byPath := make(map[string]PackageSpec, len(normalized.Packages))
	for _, spec := range normalized.Packages {
		byPath[spec.Path] = spec
	}
	image := Image{Schema: ImageSchema, InitOrder: order}
	activities := make([]EntryPlan, 0)
	for _, packagePath := range order {
		compiled, compileErr := compilePackage(byPath[packagePath])
		if compileErr != nil {
			return Image{}, compileErr
		}
		image.Packages = append(image.Packages, compiled.image)
		activities = append(activities, compiled.activities...)
	}
	entry, err := resolveEntry(normalized.Entry, byPath, activities)
	if err != nil {
		return Image{}, err
	}
	image.Entry = entry
	image.Digest = imageDigest(image)
	return image, nil
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
