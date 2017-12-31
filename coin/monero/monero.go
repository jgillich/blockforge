package monero

import (
	"errors"

	"gitlab.com/jgillich/autominer/coin"
)

func init() {
	coin.Coins["monero"] = &Monero{}
}

type Monero struct {
}

func (x *Monero) Mine(config coin.MineConfig) error {
	return errors.New("not implemented")
	//xmrstak.ExecutorInst().Ex_start(true)

}
