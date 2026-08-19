package query

import (
	"fmt"
)

func (options TraversalOptions) normalized() (TraversalOptions, error) {
	if options.MaxDepth <= 0 {
		return TraversalOptions{}, invalidTraversal("max depth must be positive")
	}
	if options.Limit < 0 || options.Limit > MaxEnvelopeLimit {
		return TraversalOptions{}, invalidTraversal(
			fmt.Sprintf("limit must be 0..%d", MaxEnvelopeLimit),
		)
	}
	if options.Direction == 0 {
		options.Direction = Outgoing
	}
	if options.Direction != Outgoing && options.Direction != Incoming && options.Direction != Both {
		return TraversalOptions{}, invalidTraversal("unknown traversal direction")
	}
	if options.Predicate != "" {
		predicate, err := ParseRelation(options.Predicate)
		if err != nil {
			return TraversalOptions{}, invalidTraversal(err.Error())
		}
		options.Predicate = predicate
	}
	selection, err := options.Selection.normalized()
	if err != nil {
		return TraversalOptions{}, err
	}
	options.Selection = selection
	return options, nil
}
func invalidTraversal(detail string) error {
	return fmt.Errorf("%w: %s", ErrInvalidTraversal, detail)
}
