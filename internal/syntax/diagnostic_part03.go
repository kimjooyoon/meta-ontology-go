package syntax

func (e diagnosticError) Error() string { return string(e) }
