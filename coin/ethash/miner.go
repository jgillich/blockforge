package ethash

import (
	"fmt"
	"net/url"

	"gitlab.com/jgillich/autominer/cgo/ethminer"
	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	ethminer ethminer.Ethminer
	config   coin.MinerConfig
}

func NewMiner(config coin.MinerConfig) (coin.Miner, error) {

	if config.Threads > 0 {
		return nil, fmt.Errorf("coin '%v' does not support cpu mining", config.Coin)
	}

	if len(config.GPUIndexes) == 0 {
		return nil, fmt.Errorf("no gpus configured for coin '%v'", config.Coin)
	}

	return &Miner{config: config}, nil
}

func (m *Miner) Start() error {
	config := m.config

	u, err := url.Parse(config.PoolURL)
	if err != nil {
		return err
	}

	openclDevices := ethminer.NewUnsignedVector(int64(len(config.GPUIndexes)))
	for _, idx := range config.GPUIndexes {
		openclDevices.Add(uint(idx))
	}

	cudaDevices := ethminer.NewUnsignedVector(int64(0))
	// TODO
	//for _, idx := range config.GPUIndexes {
	//}

	go func() {
		m.ethminer = ethminer.NewEthminer(u.Hostname(), u.Port(), config.PoolUser, config.PoolPass, config.PoolEmail, openclDevices, cudaDevices)
	}()

	return nil
}

func (m *Miner) Stop() {
	if m.ethminer != nil {
		ethminer.DeleteEthminer(m.ethminer)
		m.ethminer = nil
	}
}

func (m *Miner) Stats() coin.MinerStats {
	hashrate := 0
	if m.ethminer != nil {
		hashrate = m.ethminer.Hashrate()
	}

	return coin.MinerStats{
		Coin:     m.config.Coin,
		Hashrate: float32(hashrate),
	}
}
