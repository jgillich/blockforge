package etherum

import (
	"net/url"

	"gitlab.com/jgillich/autominer/cgo/ethminer"

	"gitlab.com/jgillich/autominer/coin"
)

func init() {
	for _, c := range []string{"eth", "etc", "exp", "ubq", "music"} {
		coin.Coins[c] = &Ethash{}
	}
}

type Ethash struct {
}

func (e *Ethash) Mine(config coin.MineConfig) error {
	if config.Threads > 0 {
		go e.mineCPU(config)
	}

	eth := ethminer.NewEthminer()

	u, err := url.Parse(config.PoolURL)
	if err != nil {
		return err
	}

	eth.SetM_farmURL(u.Hostname())
	eth.SetM_user(config.PoolUser)
	eth.SetM_pass(config.PoolPass)
	eth.SetM_port(u.Port())

	go eth.Start()

	return nil
}
