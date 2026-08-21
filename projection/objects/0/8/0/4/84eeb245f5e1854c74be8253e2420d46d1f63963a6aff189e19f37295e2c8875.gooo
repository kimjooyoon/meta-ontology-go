package semanticbinding

func unknownResult(err error) (Result, error) {
	typed, ok := err.(*Error)
	if !ok {
		typed = bindingError(CodeInput, Span{}, err.Error())
	}
	unknown := Unknown{Code: typed.Code, Message: typed.Message, Span: typed.Span, FullSuiteFallback: true}
	return Result{Status: StatusUnknown, Unknowns: []Unknown{unknown}, FullSuiteFallback: true}, typed
}
