package miner

import (
	"strings"

	"gitlab.com/jgillich/autominer/hardware/opencl"
	"gitlab.com/jgillich/autominer/hardware/processor"
	"gitlab.com/jgillich/autominer/stratum"
	"gitlab.com/jgillich/autominer/worker"
)

type Config struct {
	Version    int             `yaml:"version" json:"version"`
	Donate     int             `yaml:"donate" json:"donate"`
	Coins      map[string]Coin `yaml:"coins" json:"coins"`
	Processors []Processor     `yaml:"processors" json:"processors"`
	OpenCL     []OpenCLDevice  `yaml:"opencl" json:"opencl"`
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
	Coin      string `yaml:"coin" json:"coin"`
}

func GenerateConfig() (*Config, error) {

	config := Config{
		Version: 1,
		Donate:  5,
		Coins: map[string]Coin{
			"XMR": Coin{
				Pool: stratum.Pool{
					URL:  "stratum+tcp://xmr.coinfoundry.org:3032",
					User: "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
					Pass: "x",
				},
			},
		},
		Processors: []Processor{},
		OpenCL:     []OpenCLDevice{},
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

			hashMemSize := worker.CryptonightMemory
			computeUnits := device.CL().MaxComputeUnits()

			// 224byte extra memory is used per thread for meta data
			maxIntensity := int(device.CL().GlobalMemSize())/hashMemSize + 224

			// map intensity to a multiple of the compute unit count, 8 is the number of threads per work group
			intensity := (maxIntensity / (8 * computeUnits)) * computeUnits * 8

			// leave some free memory
			intensity--

			// TODO figure out the best maximum
			if intensity > 1000 {
				intensity = 1000
			}

			config.OpenCL = append(config.OpenCL, OpenCLDevice{
				Enable:    strings.Contains(device.Platform.Name, "Advanced Micro Devices"),
				Coin:      "XMR",
				Index:     device.Index,
				Platform:  platform.Index,
				Name:      device.Name,
				Intensity: intensity,
			})
		}
	}

	return &config, nil
}
