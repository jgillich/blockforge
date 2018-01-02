package command

import (
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/jgillich/autominer/gui"
)

func init() {
	Commands["gui"] = func() (cli.Command, error) {
		return GuiCommand{}, nil
	}
}

type GuiCommand struct {
}

func (c GuiCommand) Run(args []string) int {
	gui.Show()
	return 0
}

func (c GuiCommand) Help() string {

	helpText := `
Usage: coin miner [options]

	Launch the graphical user interface.
`
	return strings.TrimSpace(helpText)
}

func (c GuiCommand) Synopsis() string {
	return "Launch the graphical user interface"
}
