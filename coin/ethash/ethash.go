package ethash

import (
	"gitlab.com/jgillich/autominer/coin"
)

func init() {
	for _, c := range []string{"eth", "etc", "exp", "ubq", "music"} {
		coin.Coins[c] = &Ethash{}
	}
}

type Ethash struct{}

func (e *Ethash) Miner(config coin.MinerConfig) (coin.Miner, error) {
	return NewMiner(config)
}

func (e *Ethash) Info() coin.Info {
	return coin.Info{
		SupportsCPU:    false,
		SupportsOpenCL: true,
		SupportsCUDA:   true,
	}
}
