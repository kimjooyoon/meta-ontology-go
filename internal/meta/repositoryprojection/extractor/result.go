package extractor

type Result struct {
	Generated  map[string][]byte
	Paths      []string
	Operations []string
	Evidence   []StrategyEvidence
}
