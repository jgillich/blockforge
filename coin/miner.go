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
	Coin     string
	Hashrate float32
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
