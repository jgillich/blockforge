package main

//import _ "gitlab.com/jgillich/autominer/cgo"

import "gitlab.com/jgillich/autominer/currency"

func main() {
	cfg := currency.CurrencyConfig{}
	currency.Currencies["Etherum"].Mine(cfg)
}
