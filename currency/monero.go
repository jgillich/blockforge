package currency

func init() {
	Currencies["Monero"] = &Monero{}
}

type Monero struct {
}

func (x *Monero) Mine(config CurrencyConfig) {
	//xmrstak.ExecutorInst().Ex_start(true)

}
