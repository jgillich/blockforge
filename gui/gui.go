package gui

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"gitlab.com/jgillich/autominer/miner"

	rice "github.com/GeertJohan/go.rice"
)

type Gui struct {
	//view webview.WebView
}

func Show() {

	http.Handle("/", http.FileServer(rice.MustFindBox(".").HTTPBox()))

	listener, err := net.Listen("tcp", "127.0.0.1:3333")
	//listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	//go func() {
	log.Fatal(http.Serve(listener, nil))
	//}()

	/*

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

		view.Bind("miner", &Miner{config})

		view.Run()*/
}

type Miner struct {
	Config miner.Config `json:"config"`
}

func (b *Miner) Log(s string) {
	fmt.Println("log")
	fmt.Println(s)
	fmt.Println(b.Config.Donate)
}
