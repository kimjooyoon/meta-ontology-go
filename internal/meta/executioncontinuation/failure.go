package executioncontinuation

import "errors"

type Failure struct{ State, Reason string }

func (f *Failure) Error() string  { return f.State + "/" + f.Reason }
func unknown(reason string) error { return &Failure{"UNKNOWN", reason} }
func refuted(reason string) error { return &Failure{"REFUTED", reason} }
func errorState(err error) (string, string) {
	if failure, ok := errors.AsType[*Failure](err); ok {
		return failure.State, failure.Reason
	}
	return "UNKNOWN", "PROTOCOL_OBSERVATION_UNAVAILABLE"
}

func failedDispatch(r DispatchReceipt, err error) DispatchReceipt {
	state, reason := errorState(err)
	if state == "REFUTED" {
		r.State = "CONTINUATION_REFUTED"
	}
	return sealDispatch(r, reason)
}
