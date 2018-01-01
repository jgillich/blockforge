package cryptonight

import (
	"gitlab.com/jgillich/autominer/coin"
)

func init() {
	for _, c := range []string{"xmr", "etn", "itns", "sumo"} {
		coin.Coins[c] = &Cryptonight{
			light: false,
		}
	}
	for _, c := range []string{"aeon"} {
		coin.Coins[c] = &Cryptonight{
			light: true,
		}
	}
}

type Cryptonight struct {
	light bool
}

func (c *Cryptonight) Miner(config coin.MinerConfig) (coin.Miner, error) {
	return NewMiner(config, c.light)
}
