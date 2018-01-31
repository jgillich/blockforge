package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"

	"gitlab.com/blockforge/blockforge/hardware/processor"

	"gitlab.com/blockforge/blockforge/worker"

	"gopkg.in/yaml.v2"

	"github.com/gobuffalo/packr"
	"github.com/inconshreveable/mousetrap"
	"github.com/spf13/cobra"
	"github.com/zserge/webview"
	"gitlab.com/blockforge/blockforge/miner"
)

func init() {
	cmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long:  `Launch the graphical user interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		if mousetrap.StartedByExplorer() && !debug {
			hideConsoleWindow()
		}

		errors := make(chan error)
		var config miner.Config

		buf, err := ioutil.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				err := initConfig()
				if err != nil {
					go func(err error) { errors <- err }(err)
				} else {
					buf, err = ioutil.ReadFile(configPath)
					if err != nil {
						go func(err error) { errors <- err }(err)
					}
				}
			} else {
				go func(err error) { errors <- err }(err)
			}
		} else {
			err = yaml.Unmarshal(buf, &config)
			if err != nil {
				errors <- err
			}
		}

		http.Handle("/", http.FileServer(packr.NewBox("../gui")))

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			go func(err error) { errors <- err }(err)
		}
		defer listener.Close()

		go func() {
			err := http.Serve(listener, nil)
			if err != nil {
				errors <- err
			}
		}()

		processors, err := processor.GetProcessors()
		if err != nil {
			go func(err error) { errors <- err }(err)
		}

		view := webview.New(webview.Settings{
			URL:       "http://" + listener.Addr().String(),
			Title:     fmt.Sprintf("BlockForge %v", VERSION),
			Width:     1152,
			Height:    648,
			Resizable: true,
			Debug:     debug,
			ExternalInvokeCallback: func(view webview.WebView, data string) {
				if data == "__app_js_loaded__" {

					_, err := view.Bind("backend", &GuiBackend{
						errors:     errors,
						webview:    view,
						miner:      nil,
						Config:     config,
						Processors: processors,
						Coins:      worker.List(),
					})

					if err != nil {
						errors <- err
						return
					}

					if debug && runtime.GOOS == "windows" {
						view.Eval(`document.write('<script type="text/javascript" src="https://getfirebug.com/firebug-lite.js#startOpened=true"></script>')`)
					}

					err = view.Eval("init()")
					if err != nil {
						errors <- err
						return
					}

				}
			},
		})
		defer view.Exit()

		go func() {
			err := <-errors
			view.Dialog(webview.DialogTypeAlert, webview.DialogFlagError, "Unexpected Error", fmt.Sprintf("%v", err))
			os.Exit(1)
		}()

		view.Run()
	},
}

type GuiBackend struct {
	errors     chan error
	webview    webview.WebView
	miner      *miner.Miner
	Config     miner.Config                   `json:"config"`
	Processors []*processor.Processor         `json:"processors"`
	Coins      map[string]worker.Capabilities `json:"coins"`
}

func (g *GuiBackend) err(err error) {

}

func (g *GuiBackend) Start() {
	miner, err := miner.New(g.Config)
	if err != nil {
		g.errors <- err
		return
	}
	go func() {
		err := miner.Start()
		if err != nil {
			g.errors <- err
		}
	}()
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
			g.errors <- err
			return
		}
		g.webview.Eval(fmt.Sprintf("miner.trigger('stats', %v)", string(buf)))
	}
}

func (g *GuiBackend) UpdateConfig(s string) {
	var config miner.Config
	err := json.Unmarshal([]byte(s), &config)
	if err != nil {
		g.errors <- err
		return
	}

	out, err := yaml.Marshal(&config)
	if err != nil {
		g.errors <- err
		return
	}

	err = ioutil.WriteFile(configPath, []byte(out), os.ModePerm)
	if err != nil {
		g.errors <- err
		return
	}

	g.Config = config

	if g.miner != nil {
		g.Stop()
		g.Start()
	}
}
