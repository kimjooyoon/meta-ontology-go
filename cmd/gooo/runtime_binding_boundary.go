package main

import "errors"

var errRuntimeBindingsUnsupportedByGenerator = errors.New("generator: runtime bindings are unsupported")
var errRuntimePlanRequired = errors.New("generator: runtime bindings require --runtime-plan")
