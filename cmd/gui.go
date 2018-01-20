package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"gitlab.com/jgillich/autominer/hardware"

	rice "github.com/GeertJohan/go.rice"
	"github.com/hashicorp/hcl"
	"github.com/spf13/cobra"
	"github.com/zserge/webview"
	"gitlab.com/jgillich/autominer/log"
	"gitlab.com/jgillich/autominer/miner"
)

func init() {
	cmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long:  `Launch the graphical user interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		buf, err := ioutil.ReadFile(configPath)
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

		http.Handle("/", http.FileServer(rice.MustFindBox("../gui").HTTPBox()))

		listener, err := net.Listen("tcp", "127.0.0.1:3333")
		//listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		go func() {
			log.Fatal(http.Serve(listener, nil))
		}()

		view := webview.New(webview.Settings{
			URL:       "http://" + listener.Addr().String(),
			Title:     "CoinStack",
			Width:     1024,
			Height:    768,
			Resizable: true,
			Debug:     true,
		})
		defer view.Exit()

		hardware, err := hardware.New()
		if err != nil {
			log.Fatal(err)
		}

		view.Dispatch(func() {
			view.Bind("miner", &GuiMiner{view, config, hardware, nil})
			view.Eval("init()")
		})

		view.Run()
	},
}

type GuiMiner struct {
	webview  webview.WebView
	Config   miner.Config       `json:"config"`
	Hardware *hardware.Hardware `json:"hardware"`
	//Stats    []coin.MinerStats  `json:"stats"`
	miner *miner.Miner
}

func (g *GuiMiner) Start() {
	miner, err := miner.New(g.Config)
	if err != nil {
		log.Fatal(err)
	}
	g.miner = miner
}

func (g *GuiMiner) Stop() {
	if g.miner != nil {
		g.miner.Stop()
		g.miner = nil
	}
}

func (g *GuiMiner) Stats() {
	if g.miner != nil {
		buf, err := json.Marshal(g.miner.Stats())
		if err != nil {
			log.Fatal(err)
		}
		g.webview.Eval(fmt.Sprintf("window.updateStats(%v)", string(buf)))
	}
}
