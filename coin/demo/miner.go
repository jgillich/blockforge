package demo

import (
	"math/rand"

	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	config coin.MinerConfig
}

func NewMiner(config coin.MinerConfig) (coin.Miner, error) {

	return &Miner{config: config}, nil
}

func (m *Miner) Start() error {

	return nil
}

func (m *Miner) Stop() {

}

func (m *Miner) Stats() coin.MinerStats {
	var cpuStats []coin.CPUStats
	for _, cpu := range m.config.CPUSet {
		cpuStats = append(cpuStats, coin.CPUStats{
			Index:    cpu.CPU.Index,
			Hashrate: float32(100 * (rand.Intn(9) + 1)),
		})
	}

	var gpuStats []coin.GPUStats
	for _, gpu := range m.config.GPUSet {
		gpuStats = append(gpuStats, coin.GPUStats{
			Index:    gpu.GPU.Index,
			Hashrate: float32(100 * (rand.Intn(9) + 1)),
		})
	}

	return coin.MinerStats{
		GPUStats: gpuStats,
		CPUStats: cpuStats,
	}
}
