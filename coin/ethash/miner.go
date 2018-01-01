package ethash

import (
	"fmt"
	"net/url"

	"gitlab.com/jgillich/autominer/cgo/ethminer"
	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	coin     string
	ethminer ethminer.Ethminer
}

func NewMiner(config coin.MinerConfig) (coin.Miner, error) {

	miner := Miner{coin: config.Coin}

	if config.Threads > 0 {
		return nil, fmt.Errorf("coin '%v' does not support cpu mining", config.Coin)
	}

	if len(config.GPUIndexes) > 0 {
		miner.ethminer = ethminer.NewEthminer()

		u, err := url.Parse(config.PoolURL)
		if err != nil {
			return nil, err
		}

		openclDevices := make([]uint, len(config.GPUIndexes))
		for i, idx := range config.GPUIndexes {
			openclDevices[i] = uint(idx)
		}

		miner.ethminer.SetM_farmURL(u.Hostname())
		miner.ethminer.SetM_user(config.PoolUser)
		miner.ethminer.SetM_pass(config.PoolPass)
		miner.ethminer.SetM_port(u.Port())
		miner.ethminer.SetM_openclDevices(&openclDevices[0])
		miner.ethminer.SetM_openclDeviceCount(uint(len(openclDevices)))

	}

	return &miner, nil
}

func (m *Miner) Start() error {
	go m.ethminer.Start()

	return nil
}

func (m *Miner) Stats() coin.MinerStats {
	hashrate := 0
	hashrate += m.ethminer.GetM_hashrate()

	return coin.MinerStats{
		Coin:     m.coin,
		Hashrate: float32(hashrate),
	}
}
