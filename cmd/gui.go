package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"

	"gitlab.com/blockforge/blockforge/coin"
	"gitlab.com/blockforge/blockforge/log"

	"gitlab.com/blockforge/blockforge/hardware/processor"

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

		errors := make(chan error, 10)
		var config miner.Config

		buf, err := ioutil.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				err := initConfig()
				if err != nil {
					errors <- err
				} else {
					buf, err = ioutil.ReadFile(configPath)
					if err != nil {
						errors <- err
					}
				}
			} else {
				errors <- err
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
			errors <- err
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
			errors <- err
		}

		view := webview.New(webview.Settings{
			URL:       "http://" + listener.Addr().String(),
			Title:     fmt.Sprintf("BlockForge %v", VERSION),
			Width:     1232,
			Height:    700,
			Resizable: true,
			Debug:     debug,
			ExternalInvokeCallback: func(view webview.WebView, data string) {
				if data == "__app_js_loaded__" {

					_, err := view.Bind("backend", &guiBackend{
						panic:      errors,
						webview:    view,
						miner:      nil,
						Config:     config,
						Processors: processors,
						Coins:      coin.Coins,
					})

					if err != nil {
						errors <- err
						return
					}

					if debug && runtime.GOOS == "windows" {
						view.Eval(`document.write('<script type="text/javascript" src="https://getfirebug.com/firebug-lite.js"></script>')`)
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
			view.Dispatch(func() {
				view.Dialog(webview.DialogTypeAlert, webview.DialogFlagError, "Unexpected Error", fmt.Sprintf("%+v", err))
				os.Exit(1)
			})
		}()

		view.Run()
	},
}

type guiBackend struct {
	panic      chan error
	webview    webview.WebView
	miner      *miner.Miner
	Config     miner.Config           `json:"config"`
	Processors []*processor.Processor `json:"processors"`
	Coins      []coin.Coin            `json:"coins"`
	mu         sync.Mutex
}

func (g *guiBackend) Start() {
	g.mu.Lock()
	defer g.mu.Unlock()

	miner, err := miner.New(g.Config)
	if err != nil {
		g.panic <- err
		return
	}
	go func() {
		err := miner.Start()
		if err != nil {
			g.panic <- err
		}
	}()
	g.miner = miner
}

func (g *guiBackend) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.miner != nil {
		g.miner.Stop()
		g.miner = nil
	}
}

func (g *guiBackend) Stats() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.miner != nil {
		stats := g.miner.Stats()
		buf, err := json.Marshal(stats)
		if err != nil {
			log.Error(err)
			return
		}
		g.webview.Dispatch(func() {
			g.webview.Eval(fmt.Sprintf("miner.trigger('stats', %v)", string(buf)))
		})
	}
}

func (g *guiBackend) UpdateConfig(s string) {
	g.mu.Lock()
	var config miner.Config
	err := json.Unmarshal([]byte(s), &config)
	if err != nil {
		g.panic <- err
		return
	}

	out, err := yaml.Marshal(&config)
	if err != nil {
		g.panic <- err
		return
	}

	err = ioutil.WriteFile(configPath, []byte(out), os.ModePerm)
	if err != nil {
		g.panic <- err
		return
	}

	g.Config = config

	// defers are executed in last-in-first-out order (stop -> start)
	if g.miner != nil {
		defer g.Start()
		defer g.Stop()
	}
	g.mu.Unlock()
}
