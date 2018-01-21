package worker

import (
	"fmt"
	"strconv"

	"gitlab.com/jgillich/autominer/hardware"
	"gitlab.com/jgillich/autominer/stratum"
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
	Stratum stratum.Client
	Donate  int
	CPUSet  []CPUConfig
	GPUSet  []GPUConfig
}

type CPUConfig struct {
	Threads int
	CPU     hardware.CPU
}

type GPUConfig struct {
	Intensity int
	GPU       hardware.GPU
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

func hexUint32(hex []byte) uint32 {
	result := uint32(0)
	length := len(hex)
	for i := 0; i < length; i += 2 {
		d, _ := strconv.ParseInt(fmt.Sprintf("0x%v", string(hex[i:i+2])), 0, 16)
		result <<= 8
		result |= uint32(d)
	}
	return result
}

func hexUint64LE(hex []byte) uint64 {
	result := uint64(0)
	length := len(hex)
	for i := 0; i < length; i += 2 {
		d, _ := strconv.ParseInt(fmt.Sprintf("0x%v", string(hex[length-i-2:length-i])), 0, 16)
		result <<= 8
		result |= uint64(d)
	}
	return result
}

func fmtNonce(nonce uint32) string {
	return fmt.Sprintf("%08x", nonce)
}
