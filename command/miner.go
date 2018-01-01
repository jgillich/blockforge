package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/buger/goterm"

	"github.com/hashicorp/hcl"
	"gitlab.com/jgillich/autominer/coin"
	"gitlab.com/jgillich/autominer/miner"

	// import coins to initialize them
	_ "gitlab.com/jgillich/autominer/coin/cryptonight"
	_ "gitlab.com/jgillich/autominer/coin/ethash"

	"github.com/mitchellh/cli"
)

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

	miner := miner.New(config)

	stats := make(chan coin.MineStats)

	go func() {
		allStats := map[string]float32{}

		printStats := func() {
			goterm.Clear()
			goterm.MoveCursor(0, 0)
			goterm.Flush()

			for c, r := range allStats {
				fmt.Fprintf(goterm.Output, "%v: %v H/s\t", c, r)
			}

			goterm.Flush()
		}

		for stat := range stats {
			allStats[stat.Coin] = stat.Hashrate
			printStats()
		}
	}()

	err = miner.Start(stats)
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
