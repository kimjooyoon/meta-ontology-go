package cache

import "context"

// PartialSpec describes one independently reusable fragment of a projection.
// Inputs and Freshness should contain only facts that affect this fragment;
// that locality is what permits an unchanged fragment to survive a sibling
// change.
type PartialSpec struct {
	KeySpec KeySpec
	Part    string
}

// NewPartialKey scopes a content-addressed key to a stable fragment name.
// The fragment name is canonicalized with its local inputs, so sibling parts
// cannot collide even when their inputs happen to be equal.
func NewPartialKey(spec PartialSpec) (Key, error) {
	keySpec := spec.KeySpec
	if err := validateKeyComponent("partial name", spec.Part, true); err != nil {
		return Key{}, err
	}
	keySpec.Inputs = map[string]any{"inputs": spec.KeySpec.Inputs, "part": spec.Part}
	return NewKey(keySpec)
}

// PutPartial stores one independently addressable fragment.
func (c *Cache) PutPartial(spec PartialSpec, data []byte) error {
	key, err := NewPartialKey(spec)
	if err != nil {
		return err
	}
	return c.Put(key, data)
}

// GetOrComputePartial reuses one fragment without requiring sibling
// fragments to be present or current. The returned key is the exact address
// used for the lookup and store.
func (c *Cache) GetOrComputePartial(ctx context.Context, spec PartialSpec,
	compute ComputeFunc) (Key, []byte, Metadata, bool, error) {
	key, err := NewPartialKey(spec)
	if err != nil {
		return Key{}, nil, Metadata{}, false, err
	}
	data, metadata, hit, err := c.GetOrCompute(ctx, key, compute)
	return key, data, metadata, hit, err
}
