// Command shipnote is the shipnote CLI entry point.
package main

import (
	"fmt"
	"os"

	"github.com/DesmondSanctity/shipnote/cmd/shipnote/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "shipnote: "+err.Error())
		os.Exit(1)
	}
}
