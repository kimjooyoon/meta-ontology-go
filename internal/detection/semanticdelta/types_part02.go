package semanticdelta

// Violation identifies one out-of-scope endpoint, node, or predicate.
type Violation struct {
	Operation Operation  `json:"operation"`
	Change    ChangeKind `json:"change"`
	ID        string     `json:"id,omitempty"`
	Kind      string     `json:"kind,omitempty"`
	Subject   string     `json:"subject,omitempty"`
	Predicate string     `json:"predicate,omitempty"`
	Object    string     `json:"object,omitempty"`
	Endpoint  string     `json:"endpoint,omitempty"`
	Reason    string     `json:"reason"`
}

// Report is the deterministic result of scope detection.
type Report struct {
	Allowed    bool        `json:"allowed"`
	Violations []Violation `json:"violations,omitempty"`
}
