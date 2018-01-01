package etherum

import (
	"net/url"
	"time"

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
		//return errors.New("CPU mining is not supported by the ethash miner")
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

	go func() {
		for {
			config.Stats <- coin.MineStats{
				Coin:     config.Coin,
				Hashrate: float32(time.Now().Second() * 10),
			}
			time.Sleep(time.Second * 5)
		}

	}()
	return nil
}
