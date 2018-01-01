package monero

import (
	"gitlab.com/jgillich/autominer/cgo/xmrstak"
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
	xmrstak.ExecutorInst().Ex_start(true)

	return nil
}
