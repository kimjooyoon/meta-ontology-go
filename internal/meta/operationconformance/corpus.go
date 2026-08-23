package operationconformance

type ExpectedVerdict string

const (
	VerdictPass    ExpectedVerdict = "PASS"
	VerdictFail    ExpectedVerdict = "FAIL"
	VerdictUnknown ExpectedVerdict = "UNKNOWN"
)

type Fact struct {
	Key   string
	Value string
}

type CorpusCase struct {
	ID          string
	IndicatorID string
	Expected    ExpectedVerdict
	Facts       []Fact
}

func SplitGoV1Corpus() []CorpusCase {
	combined := make([]CorpusCase, 0, len(topologyCorpus)+len(semanticCorpus))
	combined = append(combined, topologyCorpus...)
	combined = append(combined, semanticCorpus...)
	result := make([]CorpusCase, len(combined))
	for index, item := range combined {
		result[index] = item
		result[index].Facts = append([]Fact(nil), item.Facts...)
	}
	return result
}
