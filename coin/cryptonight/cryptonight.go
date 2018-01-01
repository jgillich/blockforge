package monero

import (
	"errors"

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

func (x *Cryptonight) Mine(config coin.MineConfig) error {
	return errors.New("not implemented")
	//xmrstak.ExecutorInst().Ex_start(true)

}
