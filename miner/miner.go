package miner

type Miner struct {
	config Config
}

func New(config Config) *Miner {
	miner := Miner{config: config}

	return &miner
}
