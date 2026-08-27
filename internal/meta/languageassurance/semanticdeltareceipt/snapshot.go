package semanticdeltareceipt

func snapshot(raw []byte, source projectedSource, err error) Snapshot {
	result := Snapshot{SourceDigest: digestBytes(raw), Bytes: len(raw), Lines: lineCount(raw), ParseStatus: "EXACT", ParseReason: "SOURCE_PARSED"}
	if err != nil {
		result.ParseStatus, result.ParseReason = "UNKNOWN", "UNSUPPORTED_GOOO_SOURCE"
		return result
	}
	result.SemanticDigest = source.semanticDigest
	result.StructuralDigest = digestValue(structuralImage{Nodes: source.nodes, Facts: source.facts})
	result.ClaimDigest = digestValue(source.claims)
	result.Nodes, result.Facts, result.Claims = source.nodes, source.facts, source.claims
	return result
}

type structuralImage struct {
	Nodes []Node `json:"nodes"`
	Facts []Fact `json:"facts"`
}

func lineCount(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	count := 1
	for _, value := range raw {
		if value == '\n' {
			count++
		}
	}
	return count
}
