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
		go e.mineCPU(config)
	}

	eth := ethminer.NewEthminer()

	u, err := url.Parse(config.PoolURL)
	if err != nil {
		return err
	}

	openclDevices := make([]uint, len(config.GPUIndexes))
	for i, idx := range config.GPUIndexes {
		openclDevices[i] = uint(idx)
	}

	eth.SetM_farmURL(u.Hostname())
	eth.SetM_user(config.PoolUser)
	eth.SetM_pass(config.PoolPass)
	eth.SetM_port(u.Port())
	eth.SetM_openclDevices(&openclDevices[0])
	eth.SetM_openclDeviceCount(uint(len(openclDevices)))

	go eth.Start()

	go func() {
		for {
			config.Stats <- coin.MineStats{
				Coin:     config.Coin,
				Hashrate: float32(eth.GetM_hashrate()),
			}
			time.Sleep(time.Second * 10)
		}
	}()

	return nil
}
