package semanticdeltareceiptconsumer

type Verdict struct {
	Decision       string `json:"decision"`
	Resolution     string `json:"resolution"`
	Classification string `json:"classification"`
	Reason         string `json:"reason"`
	Passed         bool   `json:"passed"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
}
