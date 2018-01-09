package ethash

import (
	"net/url"

	"gitlab.com/jgillich/autominer/hardware"

	"gitlab.com/jgillich/autominer/cgo/ethminer"
	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	ethminer ethminer.Ethminer
	config   coin.MinerConfig
}

func NewMiner(config coin.MinerConfig) (coin.Miner, error) {
	return &Miner{config: config}, nil
}

func (m *Miner) Start() error {
	config := m.config

	u, err := url.Parse(config.Pool.URL)
	if err != nil {
		return err
	}

	var openclDevices []coin.GPUConfig
	for _, gpu := range config.GPUSet {
		if gpu.GPU.Backend == hardware.OpenCLBackend {
			openclDevices = append(openclDevices, gpu)
		}
	}
	openclIndexes := ethminer.NewUnsignedVector(int64(len(openclDevices)))
	for _, gpu := range openclDevices {
		openclIndexes.Add(uint(gpu.GPU.Index))
	}

	var cudaDevices []coin.GPUConfig
	for _, gpu := range config.GPUSet {
		if gpu.GPU.Backend == hardware.CUDABackend {
			cudaDevices = append(cudaDevices, gpu)
		}
	}
	cudaIndexes := ethminer.NewUnsignedVector(int64(len(cudaDevices)))
	for _, gpu := range cudaDevices {
		cudaIndexes.Add(uint(gpu.GPU.Index))
	}

	go func() {
		m.ethminer = ethminer.NewEthminer(u.Hostname(), u.Port(), config.Pool.User, config.Pool.Pass, config.Pool.Email, openclIndexes, cudaIndexes)
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
