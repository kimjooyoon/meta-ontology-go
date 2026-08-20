package main

func classify(subject inputSubject, atoms []declarationAtom) planSubject {
	result := planSubject{
		Logical: subject.Logical, Lines: subject.Value,
		RequiredSave: subject.Value - subject.Limit,
		Consumer: "logical-source-splitter", Proof: "axiomatic-foundation",
	}
	giantMovable, giantFixed := false, false
	for _, atom := range atoms {
		if atom.lines > result.MaxAtomLines {
			result.MaxAtomLines = atom.lines
		}
		if atom.movable {
			result.MovableAtoms++
			giantMovable = giantMovable || atom.lines+3 > subject.Limit
		} else {
			giantFixed = giantFixed || atom.lines+3 > subject.Limit
		}
		if atom.compactable {
			result.DensityAtoms++
		}
	}
	switch {
	case len(atoms) == 0:
		result.Reason = "unclassified"
		result.Operation = "inspect-parse-domain"
	case result.MovableAtoms == 0 && result.DensityAtoms > 0:
		result.Reason = "static-density-rewrite"
		result.Operation = "compact-static-literal"
	case result.MovableAtoms == 0:
		result.Reason = "no-movable-declaration"
		result.Operation = "extract-indivisible-source"
	case result.RequiredSave <= 10:
		result.Reason = "density-rewrite"
		result.Operation = "compact-obvious-lines"
	case result.DensityAtoms > 0:
		result.Reason = "large-density-rewrite"
		result.Operation = "compact-large-expression"
	case giantFixed:
		result.Reason = "fixed-declaration-capacity"
		result.Operation = "extract-fixed-declaration"
	case giantMovable:
		result.Reason = "movable-declaration-capacity"
		result.Operation = "extract-movable-declaration"
	default:
		result.Reason = "projectable"
		result.Operation = "split-logical-declarations"
	}
	return result
}

func indicatorCounts(subjects []planSubject) map[string]int {
	counts := map[string]int{
		"projectable": 0, "density-rewrite": 0, "static-density-rewrite": 0,
		"large-density-rewrite": 0,
		"no-movable-declaration": 0,
		"fixed-declaration-capacity": 0, "movable-declaration-capacity": 0,
		"unclassified": 0,
	}
	for _, subject := range subjects {
		counts[subject.Reason]++
	}
	return counts
}
