package main

import (
	"log"
	"os"

	"gitlab.com/jgillich/autominer/command"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("coin", "1.0.0")
	c.Args = os.Args[1:]

	ui := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr}

	c.Commands = map[string]cli.CommandFactory{
		"miner": func() (cli.Command, error) {
			return command.MinerCommand{Ui: ui}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)

}
