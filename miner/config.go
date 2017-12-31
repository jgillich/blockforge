package miner

type Config struct {
	Donate   int             `hcl:"donate"`
	LogLevel string          `hcl:"log_level"`
	Coins    map[string]Coin `hcl:"coin"`
	CPUs     map[string]CPU  `hcl:"cpu"`
	GPUs     map[string]GPU  `hcl:"gpu"`
}

type Coin struct {
	Pool Pool `hcl:"pool"`
}

type Pool struct {
	URL  string `hcl:"url"`
	User string `hcl:"user"`
	Pass string `hcl:"pass"`
}

type CPU struct {
	Name    string `hcl:"name"`
	Coin    string `hcl:"coin"`
	Threads int    `hcl:"threads"`
}

type GPU struct {
	Name string `hcl:"name"`
	Coin string `hcl:"coin"`
}
