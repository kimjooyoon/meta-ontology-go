package semantic

import "strings"

func writeCanonicalNodeAssembly(builder *strings.Builder, node Node) {
	if node.Assembly != nil {
		raw, _ := node.Assembly.Canonical()
		builder.WriteString("assembly\t")
		writeCanonicalField(builder, raw)
	}
}
