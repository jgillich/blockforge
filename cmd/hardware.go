package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/jgillich/autominer/hardware"
	"gitlab.com/jgillich/autominer/log"
)

func init() {
	cmd.AddCommand(hardwareCmd)
}

var hardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "List hardware",
	Long:  `List hardware.`,
	Run: func(cmd *cobra.Command, args []string) {
		hw, err := hardware.New()
		if err != nil {
			log.Fatal(err)
		}

		for _, cpu := range hw.CPUs {
			fmt.Printf("cpu %+v \n", cpu)

		}

		for _, gpu := range hw.GPUs {
			fmt.Printf("gpu %+v \n", gpu)
		}
	},
}
