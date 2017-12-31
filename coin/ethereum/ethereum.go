package etherum

import (
	"errors"

	"gitlab.com/jgillich/autominer/coin"
)

func init() {
	coin.Coins["ethereum"] = &Ethereum{}
}

type Ethereum struct {
}

func (e *Ethereum) Mine(config coin.MineConfig) error {
	if config.Threads > 0 {
		return errors.New("CPU mining is not supported by the ethereum miner")
	}

	return errors.New("not implemented")
}
