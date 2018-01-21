package miner

import (
	"strconv"

	"gitlab.com/jgillich/autominer/hardware"
	"gitlab.com/jgillich/autominer/stratum"
)

type Config struct {
	Donate int             `yaml:"donate" json:"donate"`
	Coins  map[string]Coin `yaml:"coins" json:"coins"`
	CPUs   map[string]CPU  `yaml:"cpus" json:"cpus"`
	GPUs   map[string]GPU  `yaml:"gpus" json:"gpus"`
}

type Coin struct {
	Pool stratum.Pool `yaml:"pool" json:"pool"`
}

type CPU struct {
	Model   string `yaml:"model" json:"model"`
	Coin    string `yaml:"coin" json:"coin"`
	Threads int    `yaml:"threads" json:"threads"`
}

type GPU struct {
	Model     string `yaml:"model" json:"model"`
	Intensity int    `yaml:"intensity" json:"intensity"`
	Coin      string `yaml:"coin" json:"coin"`
}

func GenerateConfig() (Config, error) {
	config := Config{
		Donate: 5,
		Coins: map[string]Coin{
			"XMR": Coin{
				Pool: stratum.Pool{
					URL:  "stratum+tcp://xmr.coinfoundry.org:3032",
					User: "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
					Pass: "x",
				},
			},
		},
		CPUs: map[string]CPU{},
		GPUs: map[string]GPU{},
	}

	hw, err := hardware.New()
	if err != nil {
		return config, err
	}

	for _, cpu := range hw.CPUs {
		config.CPUs[strconv.Itoa(cpu.Index)] = CPU{
			Coin:    "XMR",
			Threads: cpu.PhysicalCores,
			Model:   cpu.Model,
		}
	}

	/* TODO uncomment when GPU support is added
	for _, gpu := range hw.GPUs {
		config.GPUs[strconv.Itoa(gpu.Index)] = GPU{
			Coin:      "XMR",
			Intensity: 1,
			Model:     gpu.Model,
		}
	}
	*/

	return config, nil
}
