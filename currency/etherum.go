package currency

func init() {
	Currencies["Etherum"] = &Etherum{}
}

type Etherum struct {
}

func (x *Etherum) Mine(config CurrencyConfig) {

}
