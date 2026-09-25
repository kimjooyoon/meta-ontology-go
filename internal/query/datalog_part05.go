package query

import (
	"encoding/json"
)

// CanonicalJSON is a stable request representation after defaults and all
// bounds have been normalized. It is suitable for replay/cache keys.
func (request DatalogQuery) CanonicalJSON() ([]byte, error) {
	normalized, _, err := normalizeDatalogQuery(request)
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

// CanonicalDigest hashes the normalized Datalog request bytes.
func (request DatalogQuery) CanonicalDigest() (string, error) {
	canonical, err := request.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

// CanonicalJSON returns the deterministic result receipt. It is not a source
// or authority hash; the graph hash remains owned by the semantic snapshot.
func (result DatalogResult) CanonicalJSON() ([]byte, error) {
	return json.Marshal(result)
}

// CanonicalDigest hashes the stable Datalog result bytes.
func (result DatalogResult) CanonicalDigest() (string, error) {
	canonical, err := result.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

// QueryDatalog is an API synonym that reads naturally at call sites.
func (graph Graph) QueryDatalog(request DatalogQuery) (DatalogResult, error) {
	return graph.EvaluateDatalog(request)
}
