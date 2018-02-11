package miner

import (
	"strings"

	"gitlab.com/blockforge/blockforge/algo/cryptonight"
	"gitlab.com/blockforge/blockforge/hardware/opencl"
	"gitlab.com/blockforge/blockforge/hardware/processor"
	"gitlab.com/blockforge/blockforge/stratum"
)

var currentConfigVersion = 1

type Config struct {
	Version       int             `yaml:"version" json:"version"`
	Donate        int             `yaml:"donate" json:"donate"`
	Coins         map[string]Coin `yaml:"coins" json:"coins"`
	Processors    []Processor     `yaml:"processors" json:"processors"`
	OpenCLDevices []OpenCLDevice  `yaml:"opencl_devices" json:"opencl_devices"`
}

type Coin struct {
	Pool stratum.Pool `yaml:"pool" json:"pool"`
}

type Processor struct {
	Enable  bool   `yaml:"enable" json:"enable"`
	Index   int    `yaml:"index" json:"index"`
	Name    string `yaml:"name" json:"name"`
	Coin    string `yaml:"coin" json:"coin"`
	Threads int    `yaml:"threads" json:"threads"`
}

type OpenCLDevice struct {
	Enable    bool   `yaml:"enable" json:"enable"`
	Platform  int    `yaml:"platform" json:"platform"`
	Index     int    `yaml:"index" json:"index"`
	Name      string `yaml:"name" json:"name"`
	Intensity int    `yaml:"intensity" json:"intensity"`
	Worksize  int    `yaml:"worksize,omitempty" json:"worksize,omitempty"`
	Coin      string `yaml:"coin" json:"coin"`
}

var defaultCoinConfig = map[string]Coin{
	"XMR": Coin{
		Pool: stratum.Pool{
			URL:  "stratum+tcp://xmr.coinfoundry.org:3032",
			User: "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
			Pass: "x",
		},
	},
}

func GenerateConfig() (*Config, error) {

	config := Config{
		Version:       currentConfigVersion,
		Donate:        5,
		Coins:         defaultCoinConfig,
		Processors:    []Processor{},
		OpenCLDevices: []OpenCLDevice{},
	}

	processors, err := processor.GetProcessors()
	if err != nil {
		return nil, err
	}

	for _, processor := range processors {
		config.Processors = append(config.Processors, Processor{
			Enable:  true,
			Coin:    "XMR",
			Index:   processor.Index,
			Name:    processor.Name,
			Threads: processor.PhysicalCores,
		})
	}

	clPlatforms, err := opencl.GetPlatforms()
	if err != nil {
		return nil, err
	}

	for _, platform := range clPlatforms {
		for _, device := range platform.Devices {

			hashMemSize := cryptonight.CryptonightMemory
			computeUnits := int64(device.CL().MaxComputeUnits())

			// 224byte extra memory is used per thread for meta data
			maxIntensity := device.CL().GlobalMemSize()/int64(hashMemSize) + 224

			// map intensity to a multiple of the compute unit count, 8 is the number of threads per work group
			intensity := (maxIntensity / (8 * computeUnits)) * computeUnits * 8

			// leave some free memory
			intensity--

			// TODO figure out the best maximum
			if intensity > 1000 {
				intensity = 1000
			}

			config.OpenCLDevices = append(config.OpenCLDevices, OpenCLDevice{
				Enable:    strings.Contains(device.Platform.Name, "Advanced Micro Devices"),
				Coin:      "XMR",
				Index:     device.Index,
				Platform:  platform.Index,
				Name:      device.Name,
				Intensity: int(intensity),
			})
		}
	}

	return &config, nil
}
