package main

import (
	"os"

	"github.com/fwiedmann/prox/cmd/root"
)

func main() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
	return
}
