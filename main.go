package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"gitlab.com/jgillich/autominer/command"
	"gitlab.com/jgillich/autominer/log"
)

func main() {
	c := cli.NewCLI("coin", "1.0.0")
	c.Args = os.Args[1:]

	c.Commands = command.Commands

	debug := false
	for _, arg := range c.Args {
		fmt.Println(arg)
		if arg == "-debug" {
			debug = true
		}
	}
	log.Initialize(debug)

	exitStatus, err := c.Run()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(exitStatus)

}
