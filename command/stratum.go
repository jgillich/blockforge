package command

import (
	"strings"

	"gitlab.com/jgillich/autominer/log"
	"gitlab.com/jgillich/autominer/stratum"
	"gitlab.com/jgillich/autominer/worker"

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
		URL: "xmr.poolmining.org:3032",
		//URL:  "pool.minexmr.com:4444",
		User: "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr.coinstack",
		Pass: "x",
	}

	stratum, err := stratum.NewClient("jsonrpc", pool)
	if err != nil {
		log.Fatal(err)
	}

	config := worker.Config{
		Stratum: stratum,
	}

	worker, err := worker.New("xmr", config)
	if err != nil {
		log.Fatal(err)
	}

	err = worker.Work()
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
