package command

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	rice "github.com/GeertJohan/go.rice"
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
	fmt.Println("launching gui")

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
		Width:     800,
		Height:    600,
		Resizable: true,
		Debug:     true,
	})
	config := miner.Config{
		Donate: 1,
		CPUs: map[string]miner.CPU{
			"Intel Core i5-6200U": miner.CPU{
				Coin:    "xmr",
				Threads: 1,
			},
		},
		Coins: map[string]miner.Coin{
			"xmr": miner.Coin{
				Pool: miner.Pool{
					URL:  "stratum+tcp://xmr.poolmining.org:3032",
					User: "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
					Pass: "x",
				},
			},
		},
	}

	view.Bind("miner", &GuiMiner{config})

	view.Run()

	return 0
}

func (c GuiCommand) Help() string {

	helpText := `
Usage: coin miner [options]

	Launch the graphical user interface.
`
	return strings.TrimSpace(helpText)
}

func (c GuiCommand) Synopsis() string {
	return "Launch the graphical user interface"
}

type GuiMiner struct {
	Config miner.Config `json:"config"`
}

func (b *GuiMiner) Log(s string) {
	fmt.Println(s)
}
