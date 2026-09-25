package query

func derivedEnvelope(root ID, rule DerivedRuleID, layer Layer, depth, limit int) Request {
	return Request{
		Schema: QueryEnvelopeSchema, Operation: OperationDerived, Root: root,
		Rule: rule, Layer: layer, Direction: "outgoing", MaxDepth: depth, Limit: limit,
	}
}
func envelopeAuthorityLabel(metadata EnvelopeMetadata, view string) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == view {
			return label
		}
	}
	return AuthorityLabel{}
}
