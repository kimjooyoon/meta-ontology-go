package generator

// ToolchainIdentity is deferred because the generator does not own a build
// or receipt store. It never fabricates an environment identity.
type ToolchainIdentity struct {
	Status string `json:"status"`
	Value  string `json:"value"`
}

// ProjectionStatus records the authoritative status of this generator result.
type ProjectionStatus struct {
	Decision string   `json:"decision"`
	Refs     []string `json:"refs"`
}

// Generator renders semantic input into Go source.
type Generator struct {
	Options Options
}

// New returns a Generator configured with options.
func New(options Options) Generator {
	return Generator{Options: options}
}
