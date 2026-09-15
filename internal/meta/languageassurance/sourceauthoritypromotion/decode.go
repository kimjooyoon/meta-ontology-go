package sourceauthoritypromotion

import "encoding/json"

func decode(input Input) (assuranceDocument, upstreamDocument, bool) {
	var assurance assuranceDocument
	var upstream upstreamDocument
	if input.SubjectSHA == "" || len(input.AssuranceJSON) == 0 || len(input.UpstreamJSON) == 0 {
		return assurance, upstream, false
	}
	if json.Unmarshal(input.AssuranceJSON, &assurance) != nil {
		return assurance, upstream, false
	}
	if json.Unmarshal(input.UpstreamJSON, &upstream) != nil {
		return assurance, upstream, false
	}
	return assurance, upstream, true
}
