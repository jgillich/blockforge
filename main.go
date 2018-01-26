package main

import (
	"github.com/spf13/cobra"
	"gitlab.com/blockforge/blockforge/cmd"
)

//go:generate rice embed-go -i ./cmd -i ./worker

func main() {
	cobra.MousetrapHelpText = ""
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
