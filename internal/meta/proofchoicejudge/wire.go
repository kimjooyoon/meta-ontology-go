// Package proofchoicejudge is intentionally independent from the producer.
// It decodes the receipt wire format and recomputes the verdict from raw
// evidence, rather than importing proofchoicealgebra's evaluator.
package proofchoicejudge

type item struct {
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	Statement     string `json:"statement"`
	Choice        string `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Numerator     int    `json:"numerator,omitempty"`
	Denominator   int    `json:"denominator,omitempty"`
	Line          int    `json:"line"`
}

type transition struct {
	ClaimID       string `json:"claim_id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Choice        string `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Persistent    bool   `json:"persistent"`
	Line          int    `json:"line"`
}

type indicator struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Choice   string `json:"choice"`
	Decision string `json:"decision"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}
