package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := run(context.Background(), parseConfig()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
