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
	return errors.New("not implemented")
}
