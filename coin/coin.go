package coin

var Coins = map[string]Coin{}

type Coin interface {
	Mine(MineConfig) error
}

type CoinConfig struct{}

type MineConfig struct {
	Donate     int
	PoolURL    string
	PoolUser   string
	PoolPass   string
	Threads    int
	GPUIndexes []int
}

type MineStats struct {
}
