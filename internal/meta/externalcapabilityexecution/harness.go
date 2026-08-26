package externalcapabilityexecution

const evaluatorSource = `package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/cosmos72/gomacro/fast"
)

type result struct {
    Arithmetic string ` + "`json:\"arithmetic\"`" + `
    Function string ` + "`json:\"function\"`" + `
}

func value(interpreter *fast.Interp, source string) string {
    values, _ := interpreter.Eval(source)
    if len(values) != 1 {
        panic(fmt.Sprintf("expected one value, got %d", len(values)))
    }
    return fmt.Sprint(values[0].ReflectValue().Interface())
}

func main() {
    arithmetic := value(fast.New(), "6*7")
    interpreter := fast.New()
    interpreter.Eval("func fibonacci(n int) int { if n <= 2 { return 1 }; return fibonacci(n-1) + fibonacci(n-2) }")
    function := value(interpreter, "fibonacci(10)")
    if err := json.NewEncoder(os.Stdout).Encode(result{arithmetic, function}); err != nil {
        panic(err)
    }
}
`

type evaluatorResult struct {
	Arithmetic string `json:"arithmetic"`
	Function   string `json:"function"`
}
