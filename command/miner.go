package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/buger/goterm"

	"github.com/hashicorp/hcl"
	"gitlab.com/jgillich/autominer/coin"
	"gitlab.com/jgillich/autominer/miner"

	// import coins to initialize them
	_ "gitlab.com/jgillich/autominer/coin/cryptonight"
	_ "gitlab.com/jgillich/autominer/coin/ethash"

	"github.com/mitchellh/cli"
)

func init() {
	Commands["miner"] = func() (cli.Command, error) {
		return MinerCommand{}, nil
	}
}

type MinerCommand struct {
	Ui cli.Ui
}

func (c MinerCommand) Run(args []string) int {
	flags := flag.NewFlagSet("miner", flag.PanicOnError)
	flags.Usage = func() { c.Ui.Output(c.Help()) }

	var configPath = flags.String("config", "miner.hcl", "Config file path")

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
	err = hcl.Decode(&config, string(buf))
	if err != nil {
		fmt.Println(err)
		return 1
	}

	miner, err := miner.New(config)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	go func() {
		for {
			stats := miner.Stats()
			goterm.Clear()
			goterm.MoveCursor(0, 0)
			goterm.Flush()

			for _, stat := range stats {
				fmt.Fprintf(goterm.Output, "%v: %v H/s\t", stat.Coin, stat.Hashrate)
			}

			goterm.Flush()

			time.Sleep(time.Second * 10)
		}
	}()

	err = miner.Start()
	if err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

func (c MinerCommand) Help() string {
	coins := make([]string, 0, len(coin.Coins))
	for c := range coin.Coins {
		coins = append(coins, c)
	}
	sort.Strings(coins)

	helpText := `
Usage: coin miner [options]

	Mine coins.

	Supported coins: ` + strings.Join(coins, ", ") + `

	General Options:

		-config=<path>          Config file path.
		-init                   Generate config file.

	`
	return strings.TrimSpace(helpText)
}

func (c MinerCommand) Synopsis() string {
	return "Mine coins"
}
