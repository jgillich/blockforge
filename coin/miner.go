package coin

import (
	"gitlab.com/jgillich/autominer/hardware"
	"gitlab.com/jgillich/autominer/stratum"
)

type Miner interface {
	Start() error
	Stop()
	Stats() MinerStats
}

type MinerStats struct {
	CPUStats []CPUStats `json:"cpu_stats"`
	GPUStats []GPUStats `json:"gpu_stats"`
}

type CPUStats struct {
	Index    int     `json:"index"`
	Hashrate float32 `json:"hashrate"`
}

type GPUStats struct {
	Index    int     `json:"index"`
	Hashrate float32 `json:"hashrate"`
}

type MinerConfig struct {
	Coin   string
	Donate int
	Pool   stratum.Pool
	CPUSet []CPUConfig
	GPUSet []GPUConfig
}

type CPUConfig struct {
	Threads int
	CPU     hardware.CPU
}

type GPUConfig struct {
	Intensity int
	GPU       hardware.GPU
}
