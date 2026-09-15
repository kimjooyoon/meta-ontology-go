package semantic

// NameRef is a fully qualified display-name lookup key. It is never used as
// semantic identity and therefore cannot accidentally merge equal names from
// different namespaces.
type NameRef struct {
	Namespace Namespace
	Name      string
}

func NewNameRef(namespace Namespace, name string) (NameRef, error) {
	ns, err := ParseNamespace(namespace.String())
	if err != nil {
		return NameRef{}, err
	}
	canonicalName, err := normalizeName(name)
	if err != nil {
		return NameRef{}, err
	}
	return NameRef{Namespace: ns, Name: canonicalName}, nil
}
