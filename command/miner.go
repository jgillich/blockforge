package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	"gitlab.com/jgillich/autominer/miner"

	// import coins to initialize them
	_ "gitlab.com/jgillich/autominer/coin/etherum"
	_ "gitlab.com/jgillich/autominer/coin/monero"

	"github.com/mitchellh/cli"
)

type MinerCommand struct {
	Ui cli.Ui
}

func (c MinerCommand) Run(args []string) int {
	flags := flag.NewFlagSet("miner", flag.PanicOnError)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	var configPath = flags.String("config", "minerr.hcl", "Config file path")

	buf, err := ioutil.ReadFile(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found. Set '-config' argument or run 'coin miner -init' to generate.")
			return 1
		}
		fmt.Println(err)
		return 1
	}
	var config miner.Config
	hcl.Decode(&config, string(buf))

	fmt.Println(config.Donate)

	return 0
}

func (c MinerCommand) Help() string {
	helpText := `
Usage: coin miner [options]

	Mine coins.

	Supported coins: monero, etherum

	General Options:

		-config=<path>          Config file path.
		-init										Generate config file.

	`
	return strings.TrimSpace(helpText)
}

func (c MinerCommand) Synopsis() string {
	return "Mine coins"
}
