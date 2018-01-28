package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"

	"github.com/xlab/closer"

	"gitlab.com/blockforge/blockforge/log"
	"gitlab.com/blockforge/blockforge/miner"
	"gitlab.com/blockforge/blockforge/worker"
)

var initArg bool

func init() {
	minerCmd.PersistentFlags().BoolVar(&initArg, "init", false, "generate initial config file")

	cmd.AddCommand(minerCmd)
}

func coinList() string {
	coins := []string{}
	for name := range worker.List() {
		coins = append(coins, name)
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
				log.Fatal(err)
			}
			fmt.Printf("Wrote config file to '%v'\n", configPath)
			return
		}

		buf, err := ioutil.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal("Config file not found. Set '--config' argument or run 'coin miner --init' to generate.")
			}
			log.Fatal(err)
		}

		var config miner.Config
		err = yaml.Unmarshal(buf, &config)
		if err != nil {
			log.Fatal(err)
		}

		log.Debugf("%+v", config)

		miner, err := miner.New(config)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("miner started")

		go func() {
			for {
				time.Sleep(time.Second * 60)

				stats := miner.Stats()

				for _, stat := range stats.CPUStats {
					log.Infof("CPU %v: %.2f H/s", stat.Index, stat.Hashrate)
				}

				for _, stat := range stats.GPUStats {
					log.Infof("GPU %v/%v: %.2f H/s", stat.Platform, stat.Index, stat.Hashrate)
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
