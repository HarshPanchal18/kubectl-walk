package main

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed completion/kubectl-walk.bash
var bashCompletion string

func printCompletion() {
	fmt.Fprint(os.Stdout, bashCompletion)
}
