package cmd

import (
	"fmt"

	"gitlab.com/blockforge/blockforge/hardware/opencl"
	"gitlab.com/blockforge/blockforge/hardware/processor"
	"gitlab.com/blockforge/blockforge/log"

	"github.com/spf13/cobra"
)

func init() {
	cmd.AddCommand(hardwareCmd)
}

var hardwareCmd = &cobra.Command{
	Use:   "hardware",
	Short: "List hardware",
	Long:  `List hardware.`,
	Run: func(cmd *cobra.Command, args []string) {
		processors, err := processor.GetProcessors()
		if err != nil {
			log.Fatalf("unexpected error: %+v", err)
		}

		for _, p := range processors {
			fmt.Printf("Processor %v\n", p.Name)
		}

		platforms, err := opencl.GetPlatforms()
		if err != nil {
			log.Fatalf("unexpected error: %+v", err)
		}

		for _, p := range platforms {
			fmt.Printf("OpenCL Platform %v\n", p.Name)

			for _, d := range p.Devices {
				fmt.Printf("\tOpenCL Device %v\n", d.Name)
			}
		}
	},
}
