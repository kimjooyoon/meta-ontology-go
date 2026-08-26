package packageruntime

func Run(manifest Manifest) (Result, error) {
	image, err := Build(manifest)
	if err != nil {
		return Result{}, err
	}
	result := Result{Schema: ResultSchema, Image: image}
	for index, packagePath := range image.InitOrder {
		result.Events = append(result.Events, Event{
			Sequence: index + 1, Kind: "PACKAGE_INITIALIZED", PackagePath: packagePath,
		})
	}
	result.Events = append(result.Events, Event{
		Sequence: len(result.Events) + 1, Kind: "ACTIVITY_CONTRACT_RESOLVED",
		PackagePath: image.Entry.PackagePath, Activity: image.Entry.Activity,
	})
	result.ResultDigest = resultDigest(result)
	return result, nil
}
