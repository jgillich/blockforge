package main

import "gitlab.com/jgillich/autominer/cmd"
import 	"github.com/spf13/cobra"

func main() {
	cobra.MousetrapHelpText = ""
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
