package semantic

import (
	"fmt"
	"strings"
)

func (c InferencePathChain) Canonical() string {
	if normalized, err := NewInferencePathChain(c.Edges...); err == nil {
		c = normalized
	}
	var b strings.Builder
	b.WriteString("inference-chain\t")
	b.WriteString(fmt.Sprint(len(c.Edges)))
	b.WriteByte('\n')
	for _, edge := range c.Edges {
		b.WriteString(edge.Canonical())
		b.WriteByte('\n')
	}
	return b.String()
}
