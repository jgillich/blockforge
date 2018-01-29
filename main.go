package main

import (
	"github.com/spf13/cobra"
	"gitlab.com/blockforge/blockforge/cmd"
)

//go:generate packr -i gui/
//go:generate packr -i opencl/

func main() {
	cobra.MousetrapHelpText = ""
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
