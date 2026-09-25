package analyzer

const registrySchema = "analyzer-registry/v1"

// Registry contains semantic symbols allowed to cross the Go / semantic
// boundary. Callers provide the stable symbol keys; the registry does not load
// packages or infer types.
type Registry struct {
	entries []Registration
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a semantic symbol. Multiple identities for one Go reference
// are intentionally retained as ambiguity during analysis.
func (r *Registry) Register(registration Registration) error {
	if r == nil {
		return errNilRegistry
	}
	if err := validateRegistryRegistration(registration); err != nil {
		return err
	}
	for _, existing := range r.entries {
		if sameRegistration(existing, registration) {
			return nil
		}
	}
	r.entries = append(r.entries, registration)
	return nil
}

// MustRegister is Register for package initialization and tests.
func (r *Registry) MustRegister(registration Registration) {
	if err := r.Register(registration); err != nil {
		panic(err)
	}
}

// RegisterActivity registers an Activity reference.
func (r *Registry) RegisterActivity(ref SymbolRef, identity Identity) error {
	return r.Register(Registration{Ref: ref, Kind: KindActivity, Identity: identity})
}

// RegisterEntity registers an Entity reference.
func (r *Registry) RegisterEntity(ref SymbolRef, identity Identity) error {
	return r.Register(Registration{Ref: ref, Kind: KindEntity, Identity: identity})
}
func (r *Registry) all() []Registration {
	if r == nil {
		return nil
	}
	entries := append([]Registration(nil), r.entries...)
	sortRegistrations(entries)
	return entries
}
