package valueexecution

// RuntimeBindingCount reports compiled runtime binding metadata only. It does
// not grant an artifact execution, adoption, or write authority.
func (plan Plan) RuntimeBindingCount() int {
	return len(plan.bindings) + len(plan.feedbacks)
}
