package monero

func init() {
	coin.Coins["Monero"] = &Monero{}
}

type Monero struct {
}

func (x *Monero) Mine(config CoinConfig) {
	//xmrstak.ExecutorInst().Ex_start(true)

}
