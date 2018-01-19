package command

import (
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/jgillich/autominer/hardware"
	"gitlab.com/jgillich/autominer/log"
)

func init() {
	Commands["hardware"] = func() (cli.Command, error) {
		return HardwareCommand{}, nil
	}
}

type HardwareCommand struct{}

func (c HardwareCommand) Run(args []string) int {
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

	return 0
}

func (c HardwareCommand) Help() string {
	helpText := `
Usage: coin hardware [options]

	List hardware.
	`
	return strings.TrimSpace(helpText)
}

func (c HardwareCommand) Synopsis() string {
	return "List hardware"
}
