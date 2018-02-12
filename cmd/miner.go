package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"gitlab.com/blockforge/blockforge/coin"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"

	"github.com/xlab/closer"

	"gitlab.com/blockforge/blockforge/log"
	"gitlab.com/blockforge/blockforge/miner"
)

var initArg bool

func init() {
	minerCmd.PersistentFlags().BoolVar(&initArg, "init", false, "generate initial config file")

	cmd.AddCommand(minerCmd)
}

func coinList() string {
	coins := []string{}
	for _, coin := range coin.Coins {
		coins = append(coins, fmt.Sprintf("%v (%v)", coin.LongName, coin.ShortName))
	}
	sort.Strings(coins)
	return strings.Join(coins, ", ")
}

var minerCmd = &cobra.Command{
	Use:   "miner",
	Short: "Mine coins",
	Long: strings.TrimSpace(`
Mine coins.

Supported coins: ` + coinList()),
	Run: func(cmd *cobra.Command, args []string) {
		if initArg {
			err := initConfig()
			if err != nil {
				log.Fatalf("unexpected error: %+v", err)
			}
			fmt.Printf("Wrote config file to '%v'\n", configPath)
			return
		}

		buf, err := ioutil.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal("Config file not found. Set '--config' argument or run 'coin miner --init' to generate.")
			}
			log.Fatalf("unexpected error: %+v", err)
		}

		var config miner.Config
		err = yaml.Unmarshal(buf, &config)
		if err != nil {
			log.Fatalf("unexpected error: %+v", err)
		}

		miner, err := miner.New(config)
		if err != nil {
			log.Fatalf("unexpected error: %+v", err)
		}

		go func() {
			err := miner.Start()
			if err != nil {
				log.Fatalf("unexpected error: %+v", err)
			}
		}()

		log.Info("miner started")

		go func() {
			for {
				time.Sleep(time.Second * 60)
				for key, hps := range miner.Stats() {
					log.Infof("%v: %.2f H/s", key, hps)
				}
			}
		}()

		closer.Bind(func() {
			miner.Stop()
		})

		// hodl
		closer.Hold()
	},
}
