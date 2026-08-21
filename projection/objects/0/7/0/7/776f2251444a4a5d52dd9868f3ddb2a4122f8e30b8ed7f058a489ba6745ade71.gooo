package analyzer

func (r *resolver) lookup(ref SymbolRef) []Registration {
	if ref.Name == "" {
		return nil
	}
	all := r.allRegistrations()
	var exact []Registration
	for _, entry := range all {
		if !sameSymbol(entry.Ref, ref) || ref.PackagePath == "" {
			continue
		}
		if entry.Ref.PackagePath == ref.PackagePath {
			exact = append(exact, entry)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	var packageName []Registration
	for _, entry := range all {
		if !sameSymbol(entry.Ref, ref) || ref.PackageName == "" {
			continue
		}
		if entry.Ref.PackageName == ref.PackageName && (entry.Ref.PackagePath == "" || ref.PackagePath == "") {
			packageName = append(packageName, entry)
		}
	}
	return packageName
}
func sameSymbol(left, right SymbolRef) bool {
	return left.Name == right.Name && left.Receiver == right.Receiver
}
func (r *resolver) allRegistrations() []Registration {
	all := make([]Registration, 0, len(r.registry)+len(r.locals))
	all = append(all, r.registry...)
	all = append(all, r.locals...)
	return uniqueRegistrations(all)
}
func makeResolution(entries []Registration) resolution {
	switch len(entries) {
	case 0:
		return resolution{state: unresolved}
	case 1:
		return resolution{state: resolved, entries: entries}
	default:
		return resolution{state: ambiguous, entries: entries}
	}
}
