package worker

import (
	metrics "github.com/armon/go-metrics"
	"gitlab.com/blockforge/blockforge/hardware/opencl"
	"gitlab.com/blockforge/blockforge/hardware/processor"
)

type Worker interface {
	Configure(Config) error
	Start() error
	Capabilities() Capabilities
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
	Metrics    *metrics.Metrics
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
