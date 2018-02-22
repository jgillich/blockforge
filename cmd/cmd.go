package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/inconshreveable/mousetrap"
	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"gitlab.com/blockforge/blockforge/log"
	"gitlab.com/blockforge/blockforge/miner"
)

var VERSION = "devel"

func init() {
	cmd.Version = VERSION
}

var (
	debug      bool
	configPath string
)

var cmd = &cobra.Command{
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Initialize(debug)
	},
	Use: "blockforge",
	Long: strings.TrimSpace(`
BlockForge is a next generation miner for cryptocurrencies.
Easy to use, multi algo and open source.
		`),
	Run: func(cmd *cobra.Command, args []string) {
		if mousetrap.StartedByExplorer() {
			guiCmd.Run(cmd, args)
		} else {
			cmd.Help()
		}
	},
}

func init() {
	configDirs := configdir.New("", "blockforge")
	defaultPath := path.Join(configDirs.QueryFolders(configdir.Global)[0].Path, "config.yml")

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	cmd.PersistentFlags().StringVar(&configPath, "config", defaultPath, "config file path")
}

func Execute() error {
	return cmd.Execute()
}

func initConfig() (*miner.Config, error) {
	config, err := miner.GenerateConfig()
	if err != nil {
		return nil, err
	}

	out, err := yaml.Marshal(&config)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
	if err != nil {
		return nil, err
	}

	return config, ioutil.WriteFile(configPath, out, os.ModePerm)
}

func updateConfig(config *miner.Config) error {
	if err := miner.UpdateConfig(config); err != nil {
		return err
	}

	out, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath, out, os.ModePerm)
}
