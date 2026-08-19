package semantic

import (
	"sort"
)

func (r TypeRegistry) lookupName(namespace Namespace, name string) []ID {
	if namespace != "" {
		ref, err := NewNameRef(namespace, name)
		if err != nil {
			return nil
		}
		return append([]ID(nil), r.names[ref]...)
	}
	ids := make([]ID, 0)
	for ref, candidates := range r.names {
		if ref.Name != name {
			continue
		}
		ids = append(ids, candidates...)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	unique := ids[:0]
	for _, id := range ids {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	return unique
}
