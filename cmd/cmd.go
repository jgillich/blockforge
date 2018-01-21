package cmd

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/inconshreveable/mousetrap"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"gitlab.com/jgillich/autominer/log"
	"gitlab.com/jgillich/autominer/miner"
)

var (
	debug      bool
	configPath string
)

var cmd = &cobra.Command{
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Initialize(debug)
	},
	Use:   "coinstack",
	Short: "CoinStack is a miner for cryptocurrencies",
	Long: strings.TrimSpace(`
CoinStack is a next generation miner for many cryptocurrencies
that features automatic hardware detection and a optional
graphical user interface.`),
	Run: func(cmd *cobra.Command, args []string) {
		if mousetrap.StartedByExplorer() {
			guiCmd.Run(cmd, args)
		} else {
			cmd.Help()
		}
	},
}

func init() {
	configDirs := configdir.New("", "coinstack")
	defaultPath := path.Join(configDirs.QueryFolders(configdir.Global)[0].Path, "config.yml")

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	cmd.PersistentFlags().StringVar(&configPath, "config", defaultPath, "config file path")
}

func Execute() error {
	return cmd.Execute()
}

func initConfig() error {
	config, err := miner.GenerateConfig()
	if err != nil {
		return err
	}

	out, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}

	_, err = file.Write(out)
	if err != nil {
		return err
	}

	return file.Close()
}
