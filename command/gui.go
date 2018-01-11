package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"gitlab.com/jgillich/autominer/hardware"

	rice "github.com/GeertJohan/go.rice"
	"github.com/hashicorp/hcl"
	"github.com/mitchellh/cli"
	"github.com/zserge/webview"
	"gitlab.com/jgillich/autominer/miner"
)

func init() {
	Commands["gui"] = func() (cli.Command, error) {
		return GuiCommand{}, nil
	}
}

type GuiCommand struct {
}

func (c GuiCommand) Run(args []string) int {
	flags := flag.NewFlagSet("miner", flag.PanicOnError)
	flags.Usage = func() { ui.Output(c.Help()) }

	var configPath = flags.String("config", "miner.hcl", "Config file path")

	buf, err := ioutil.ReadFile(*configPath)
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

	return 0
}

func (c GuiCommand) Help() string {

	helpText := `
Usage: coin miner [options]

	Launch the graphical user interface.

	General Options:

	-config=<path>          Config file path.
`
	return strings.TrimSpace(helpText)
}

func (c GuiCommand) Synopsis() string {
	return "Launch the graphical user interface"
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
	err = miner.Start()
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
