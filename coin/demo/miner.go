package demo

import (
	"math/rand"

	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	config coin.MinerConfig
}

func NewMiner(config coin.MinerConfig) (coin.Miner, error) {

	return &Miner{config: config}, nil
}

func (m *Miner) Start() error {

	return nil
}

func (m *Miner) Stop() {

}

func (m *Miner) Stats() coin.MinerStats {
	return coin.MinerStats{
		Coin:     m.config.Coin,
		Hashrate: float32(rand.Intn(100) * 10),
	}
}
