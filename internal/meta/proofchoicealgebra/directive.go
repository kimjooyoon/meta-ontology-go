package proofchoicealgebra

type directive struct {
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	ClaimID       string `json:"claim_id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Statement     string `json:"statement"`
	Choice        Choice `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Persistent    bool   `json:"persistent"`
}
