package worker

import (
	"fmt"

	"gitlab.com/blockforge/blockforge/hardware/opencl"
	"gitlab.com/blockforge/blockforge/hardware/processor"

	"gitlab.com/blockforge/blockforge/stratum"
)

var workers = map[string]workerFactory{}

type workerFactory func(Config) Worker

func New(coin string, config Config) (Worker, error) {
	factory, ok := workers[coin]
	if !ok {
		return nil, fmt.Errorf("worker for coin '%v' does not exist", coin)
	}

	return factory(config), nil
}

func List() map[string]Capabilities {
	list := map[string]Capabilities{}

	for name, factory := range workers {
		list[name] = factory(Config{}).Capabilities()
	}

	return list
}

type Worker interface {
	Work() error
	Capabilities() Capabilities
	Stats() Stats
}

type Capabilities struct {
	CPU    bool `json:"cpu"`
	OpenCL bool `json:"opencl"`
	CUDA   bool `json:"cuda"`
}

type Config struct {
	Stratum    stratum.Client
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
	Index    int     `json:"index"`
	Hashrate float32 `json:"hashrate"`
}
