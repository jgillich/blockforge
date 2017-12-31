package monero

import "gitlab.com/jgillich/autominer/coin"

func init() {
	coin.Coins["Monero"] = &Monero{}
}

type Monero struct {
}

func (x *Monero) Mine(config coin.CoinConfig) {
	//xmrstak.ExecutorInst().Ex_start(true)

}
