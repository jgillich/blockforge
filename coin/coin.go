package coin

var Coins = map[string]Coin{}

type Coin interface {
	Miner(MinerConfig) (Miner, error)
	//Info() CoinInfo
}

type Miner interface {
	Start() error
	Stats() MinerStats
}

/*
type CoinInfo struct {
	SupportsCPU    bool
	SupportsOpenGL bool
	SupportsCUDA   bool
}
*/

type CoinConfig struct{}

type MinerConfig struct {
	Coin       string
	Donate     int
	PoolURL    string
	PoolUser   string
	PoolPass   string
	Threads    int
	GPUIndexes []int
}

type MinerStats struct {
	Coin     string
	Hashrate float32
}
