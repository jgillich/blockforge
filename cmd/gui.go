package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"gitlab.com/jgillich/autominer/hardware/processor"

	"gitlab.com/jgillich/autominer/worker"

	"gopkg.in/yaml.v2"

	rice "github.com/GeertJohan/go.rice"
	"github.com/inconshreveable/mousetrap"
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
		if mousetrap.StartedByExplorer() {
			hideConsoleWindow()
		}

		buf, err := ioutil.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				err := initConfig()
				if err != nil {
					log.Fatal(err)
				}
				buf, err = ioutil.ReadFile(configPath)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}

		var config miner.Config
		err = yaml.Unmarshal(buf, &config)
		if err != nil {
			log.Fatal(err)
		}

		http.Handle("/", http.FileServer(rice.MustFindBox("../gui").HTTPBox()))

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		go func() {
			log.Fatal(http.Serve(listener, nil))
		}()

		processors, err := processor.GetProcessors()
		if err != nil {
			log.Fatal(err)
		}

		view := webview.New(webview.Settings{
			URL:       "http://" + listener.Addr().String(),
			Title:     "CoinStack",
			Width:     1024,
			Height:    768,
			Resizable: true,
			Debug:     debug,
			ExternalInvokeCallback: func(view webview.WebView, data string) {
				if data == "__app_js_loaded__" {

					view.Bind("backend", &GuiBackend{
						webview:    view,
						miner:      nil,
						Config:     config,
						Processors: processors,
						Coins:      worker.List(),
					})

					view.Eval("init()")
				}
			},
		})
		defer view.Exit()

		view.Run()
	},
}

type GuiBackend struct {
	webview    webview.WebView
	miner      *miner.Miner
	Config     miner.Config                   `json:"config"`
	Processors []processor.Processor          `json:"processors"`
	Coins      map[string]worker.Capabilities `json:"coins"`
}

func (g *GuiBackend) Start() {
	miner, err := miner.New(g.Config)
	if err != nil {
		log.Fatal(err)
	}
	g.miner = miner
}

func (g *GuiBackend) Stop() {
	if g.miner != nil {
		g.miner.Stop()
		g.miner = nil
	}
}

func (g *GuiBackend) Stats() {
	if g.miner != nil {
		buf, err := json.Marshal(g.miner.Stats())
		if err != nil {
			log.Fatal(err)
		}
		g.webview.Eval(fmt.Sprintf("miner.trigger('stats', %v)", string(buf)))
	}
}

func (g *GuiBackend) UpdateConfig(s string) {
	var config miner.Config
	err := json.Unmarshal([]byte(s), &config)
	if err != nil {
		log.Fatal(err)
	}

	out, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(configPath, []byte(out), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	g.Config = config

	if g.miner != nil {
		g.Stop()
		g.Start()
	}
}
