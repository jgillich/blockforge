package main

import (
	"github.com/spf13/cobra"
	"gitlab.com/jgillich/autominer/cmd"
)

func main() {
	cobra.MousetrapHelpText = ""
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
