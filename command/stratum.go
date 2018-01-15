package command

import (
	"log"
	"strings"

	"gitlab.com/jgillich/autominer/stratum"

	"github.com/mitchellh/cli"
)

func init() {
	Commands["stratum"] = func() (cli.Command, error) {
		return StratumCommand{}, nil
	}
}

type StratumCommand struct{}

func (c StratumCommand) Run(args []string) int {

	pool := stratum.Pool{
		URL:  "xmr.poolmining.org:3032",
		User: "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
		Pass: "x",
	}

	client, err := stratum.NewClient(pool)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	return 0
}

func (c StratumCommand) Help() string {
	helpText := `
Usage: coin stratum [options]

	Debug stratum.
	`
	return strings.TrimSpace(helpText)
}

func (c StratumCommand) Synopsis() string {
	return "Debug stratum"
}
