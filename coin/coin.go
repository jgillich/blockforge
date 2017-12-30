package coin

// TODO how to populate Coins without import cycle?
var Coins = map[string]Coin{}

type Coin interface {
	Mine(CoinConfig)
}

type CoinConfig struct{}
