package worker

import (
	"gitlab.com/blockforge/blockforge/hardware/opencl"
	"gitlab.com/blockforge/blockforge/hardware/processor"
)

type Worker interface {
	Configure(Config) error
	Start() error
	Capabilities() Capabilities
	Stats() Stats
}

type Capabilities struct {
	CPU    bool `json:"cpu"`
	OpenCL bool `json:"opencl"`
	CUDA   bool `json:"cuda"`
}

type Config struct {
	Donate     int
	Processors []ProcessorConfig
	CLDevices  []CLDeviceConfig
}

type ProcessorConfig struct {
	Threads   int
	Processor *processor.Processor
}

type CLDeviceConfig struct {
	Intensity int
	Worksize  int
	Device    *opencl.Device
}

type Stats struct {
	CPUStats []CPUStats `json:"cpu_stats"`
	GPUStats []GPUStats `json:"gpu_stats"`
}

type CPUStats struct {
	Index    int     `json:"index"`
	Hashrate float32 `json:"hashrate"`
}

type GPUStats struct {
	Platform int     `json:"platform"`
	Index    int     `json:"index"`
	Hashrate float32 `json:"hashrate"`
}
