package cmd

import (
	"path"

	"github.com/shibukawa/configdir"
	"github.com/spf13/cobra"
	"gitlab.com/jgillich/autominer/log"
)

var (
	debug      bool
	configPath string
)

var cmd = &cobra.Command{
	Use:   "coinstack",
	Short: "CoinStack is a miner for cryptocurrencies",
	Long: `
CoinStack is a next generation miner for many cryptocurrencies
and features automatic hardware detection and a optional
graphical user interface.
								`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	configDirs := configdir.New("", "coinstack")
	defaultPath := path.Join(configDirs.QueryFolders(configdir.Global)[0].Path, "config.hcl")

	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	cmd.PersistentFlags().StringVar(&configPath, "config", defaultPath, "config file path")

	log.Initialize(debug)
}

func Execute() error {
	return cmd.Execute()
}
