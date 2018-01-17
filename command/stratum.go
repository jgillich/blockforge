package command

import (
	"log"
	"strings"

	"gitlab.com/jgillich/autominer/worker"

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
		//URL: "xmr.poolmining.org:3032",
		URL:  "pool.minexmr.com:4444",
		User: "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
		Pass: "x",
	}

	stratum, err := stratum.NewClient(pool)
	if err != nil {
		log.Fatal(err)
	}

	err = stratum.Connect()
	if err != nil {
		log.Fatal(err)
	}

	worker := worker.NewMoneroWorker(stratum)

	for {
		err := worker.Work()
		if err != nil {
			log.Fatal(err)
		}
	}
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
