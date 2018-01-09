package miner

import "gitlab.com/jgillich/autominer/stratum"

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
	Coin    string `hcl:"coin" json:"coin"`
	Threads int    `hcl:"threads" json:"threads"`
}

type GPU struct {
	Intensity int    `hcl:"intensity" json:"intensity"`
	Coin      string `hcl:"coin" json:"coin"`
}
