package main

import (
	"fmt"
	"os"

	raven "github.com/getsentry/raven-go"
	"github.com/spf13/cobra"
	"gitlab.com/blockforge/blockforge/cmd"
	"gitlab.com/blockforge/blockforge/log"
)

//go:generate packr

var (
	VERSION = "devel"
	DSN     = ""
)

func main() {
	raven.SetDSN(DSN)
	cobra.MousetrapHelpText = ""

	err, _ := raven.CapturePanicAndWait(func() {
		err := cmd.Execute(VERSION)
		if err != nil {
			log.Panic(err)
		}
	}, nil)

	if err != nil {
		fmt.Printf("fatal error: %v\n", err)
		os.Exit(1)
	}
}
