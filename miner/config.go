package miner

import (
	"strconv"

	"gitlab.com/jgillich/autominer/hardware"
	"gitlab.com/jgillich/autominer/stratum"
)

type Config struct {
	Donate   int             `hcl:"donate" json:"donate"`
	LogLevel string          `hcl:"log_level" json:"log_level"`
	Coins    map[string]Coin `hcl:"coin" json:"coin"`
	CPUs     map[string]CPU  `hcl:"cpu" json:"cpu"`
	GPUs     map[string]GPU  `hcl:"gpu" json:"gpu"`
}

type Coin struct {
	Pool stratum.Pool `hcl:"pool" json:"pool"`
}

type CPU struct {
	Model   string `hcl:"model" json:"model"`
	Coin    string `hcl:"coin" json:"coin"`
	Threads int    `hcl:"threads" json:"threads"`
}

type GPU struct {
	Model     string `hcl:"model" json:"model"`
	Intensity int    `hcl:"intensity" json:"intensity"`
	Coin      string `hcl:"coin" json:"coin"`
}

func GenerateConfig() (Config, error) {
	config := Config{
		Donate: 5,
		Coins: map[string]Coin{
			"xmr": Coin{
				Pool: stratum.Pool{
					URL:  "stratum+tcp://xmr.poolmining.org:3032",
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
			Coin:    "xmr",
			Threads: cpu.PhysicalCores - 1,
			Model:   cpu.Model,
		}
	}

	for _, gpu := range hw.GPUs {
		config.GPUs[strconv.Itoa(gpu.Index)] = GPU{
			Coin:      "xmr",
			Intensity: 1,
			Model:     gpu.Model,
		}
	}

	return config, nil
}
