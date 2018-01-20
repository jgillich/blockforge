package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	jsonParser "github.com/hashicorp/hcl/json/parser"
	"github.com/spf13/cobra"

	"github.com/hashicorp/hcl/hcl/printer"

	"github.com/xlab/closer"

	"github.com/hashicorp/hcl"
	"gitlab.com/jgillich/autominer/log"
	"gitlab.com/jgillich/autominer/miner"
	"gitlab.com/jgillich/autominer/worker"
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

			err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}

			file, err := os.Create(configPath)
			if err != nil {
				log.Fatal(err)
			}

			err = printer.Fprint(file, ast)
			if err != nil {
				log.Fatal(err)
			}

			file.Close()
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
		err = hcl.Decode(&config, string(buf))
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
					log.Infof("CPU %v: %v H/s", stat.Index, stat.Hashrate)
				}

				for _, stat := range stats.GPUStats {
					log.Infof("GPU %v: %v H/s", stat.Index, stat.Hashrate)
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
