package demo

import (
	"gitlab.com/jgillich/autominer/coin"
)

func init() {
	for _, c := range []string{"demo"} {
		coin.Coins[c] = &Demo{}
	}
}

type Demo struct {
}

func (c *Demo) Miner(config coin.MinerConfig) (coin.Miner, error) {
	return NewMiner(config)
}
