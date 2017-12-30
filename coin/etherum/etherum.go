package etherum

import (
	"gitlab.com/jgillich/autominer/coin"
)

func init() {
	coin.Coins["Ethereum"] = &Ethereum{}
}

type Ethereum struct {
}

func (e *Ethereum) Mine(config coin.CoinConfig) {

}
