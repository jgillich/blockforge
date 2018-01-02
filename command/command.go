package command

import (
	"os"

	"github.com/mitchellh/cli"
)

var (
	Commands = map[string]cli.CommandFactory{}
	ui       = &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr}
)
