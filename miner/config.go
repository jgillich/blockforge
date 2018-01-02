package miner

type Config struct {
	Donate   int             `hcl:"donate" json:"donate"`
	LogLevel string          `hcl:"log_level" json:"log_level"`
	Coins    map[string]Coin `hcl:"coin" json:"coin"`
	CPUs     map[string]CPU  `hcl:"cpu" json:"cpu"`
	GPUs     map[string]GPU  `hcl:"gpu" json:"gpu"`
}

type Coin struct {
	Pool Pool `hcl:"pool" json:"pool"`
}

type Pool struct {
	URL  string `hcl:"url" json:"url"`
	User string `hcl:"user" json:"user"`
	Pass string `hcl:"pass" json:"pass"`
}

type CPU struct {
	Coin    string `hcl:"coin" json:"coin"`
	Threads int    `hcl:"threads" json:"threads"`
}

type GPU struct {
	Index int    `hcl:"index" json:"index"`
	Coin  string `hcl:"coin" json:"coin"`
}
