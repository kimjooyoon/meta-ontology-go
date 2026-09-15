package shadow

func injectionProductionVector() productionOutput {
	value := baselineProductionVector()
	value.BaseSourceDigest = "sha256:a7833e9907231e0cac60514be4fe280da6965857e7bcda68293bd379e14b8d70"
	value.HeadSourceDigest = "sha256:30f950d9a5f8eaabf50ac914469687b4ce20d78622d224d5f9d4b93da7ddfb39"
	value.RegistryDigest = "989f8295246f04651fa738e763e372c8f21e56dc0672cb902e71961d180a3562"
	value.SelectedCommands = []productionCommand{{ID: "urn:gooo:shadow/command/test", Argv: []string{"sh", "-c", "echo SAFE; touch /tmp/gooo-shadow-must-not-run"}}}
	value.Lane.RegistryDigest = value.RegistryDigest
	value.CanonicalDigest = "b5703c6d2c120cbc9750b1e37cd6f5c0b5732a683780e27514429f08c88012a5"
	return value
}
