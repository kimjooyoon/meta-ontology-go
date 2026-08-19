package metarecognition

import (
	"encoding/json"
	"sort"
)

func (r Reason) Valid() bool {
	for _, value := range []Reason{ReasonExactBinding, ReasonRenameBinding, ReasonBlobWithoutID,
		ReasonSourceMapRegistry, ReasonUnknownGraph, ReasonMissedObligation, ReasonGlobalGuard,
		ReasonSelectedDrift, ReasonOmittedFailure, ReasonNonAuthoritative, ReasonDuplicateReceipt,
		ReasonConflictingReceipt, ReasonInvalidResource, ReasonExternalMissing} {
		if r == value {
			return true
		}
	}
	return false
}
func (c Case) normalized() Case {
	c.Expected.LocalizedIDs = sorted(c.Expected.LocalizedIDs)
	c.Baseline.UnknownIDs = sorted(c.Baseline.UnknownIDs)
	c.Baseline.MissedIDs = sorted(c.Baseline.MissedIDs)
	c.Baseline.Roots = sorted(c.Baseline.Roots)
	c.Baseline.Path.IDs = sorted(c.Baseline.Path.IDs)
	c.Baseline.Commands = append([]CommandAssertion(nil), c.Baseline.Commands...)
	sort.Slice(c.Baseline.Commands, func(i, j int) bool { return c.Baseline.Commands[i].ID < c.Baseline.Commands[j].ID })
	return c
}
func (m Manifest) CanonicalJSON() ([]byte, error) {
	cases := append([]ComparisonCase(nil), m.Cases...)
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	copyManifest := m
	copyManifest.Cases = cases
	return json.Marshal(copyManifest)
}
func sorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
