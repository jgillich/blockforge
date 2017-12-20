package currency

var Currencies = map[string]Currency{}

type Currency interface {
	Mine(CurrencyConfig)
}

type CurrencyConfig struct{}
