package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	jsonParser "github.com/hashicorp/hcl/json/parser"

	"github.com/hashicorp/hcl/hcl/printer"

	"github.com/buger/goterm"
	"github.com/xlab/closer"

	"github.com/hashicorp/hcl"
	"gitlab.com/jgillich/autominer/miner"
	"gitlab.com/jgillich/autominer/worker"

	"github.com/mitchellh/cli"
)

func init() {
	Commands["miner"] = func() (cli.Command, error) {
		return MinerCommand{}, nil
	}
}

type MinerCommand struct{}

func (c MinerCommand) Run(args []string) int {
	flags := flag.NewFlagSet("miner", flag.PanicOnError)
	flags.Usage = func() { ui.Output(c.Help()) }

	var configPath = flags.String("config", "miner.hcl", "Config file path")
	var init = flags.Bool("init", false, "Generate config file")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if *init {
		config, err := miner.GenerateConfig()
		if err != nil {
			log.Fatal(err)
		}

		// TODO find a better way do serialize to hcl
		json, err := json.Marshal(config)
		if err != nil {
			log.Fatal(err)
		}
		ast, err := jsonParser.Parse(json)

		file, err := os.Create(*configPath)
		if err != nil {
			log.Fatal(err)
		}

		err = printer.Fprint(file, ast)
		if err != nil {
			log.Fatal(err)
		}

		file.Close()
		fmt.Printf("Wrote config file to '%v'\n", *configPath)
		return 0
	}

	buf, err := ioutil.ReadFile(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("Config file not found. Set '-config' argument or run 'coin miner -init' to generate.")
		}
		log.Fatal(err)
	}

	var config miner.Config
	err = hcl.Decode(&config, string(buf))
	if err != nil {
		log.Fatal(err)
	}

	miner, err := miner.New(config)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			stats := miner.Stats()
			goterm.Clear()
			goterm.MoveCursor(0, 0)
			goterm.Flush()

			for _, stat := range stats.CPUStats {
				fmt.Fprintf(goterm.Output, "CPU %v: %v H/s\n", stat.Index, stat.Hashrate)
			}

			for _, stat := range stats.GPUStats {
				fmt.Fprintf(goterm.Output, "GPU %v: %v H/s\n", stat.Index, stat.Hashrate)
			}

			goterm.Flush()

			time.Sleep(time.Second * 2)
		}
	}()

	err = miner.Start()
	if err != nil {
		log.Fatal(err)
	}

	closer.Bind(func() {
		miner.Stop()
	})

	// hodl
	closer.Hold()

	return 0
}

func (c MinerCommand) Help() string {
	coins := []string{}
	for name := range worker.List() {
		coins = append(coins, name)
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
