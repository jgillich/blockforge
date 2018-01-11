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
	Coin     string  `json:"coin"`
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
