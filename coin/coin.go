package coin

var Coins = map[string]Coin{}

type Coin interface {
	Mine(MineConfig) error
	//Info() CoinInfo
}

/*
type CoinInfo struct {
	SupportsCPU    bool
	SupportsOpenGL bool
	SupportsCUDA   bool
}
*/

type CoinConfig struct{}

type MineConfig struct {
	Coin       string
	Donate     int
	PoolURL    string
	PoolUser   string
	PoolPass   string
	Threads    int
	GPUIndexes []int
}

type MineStats struct {
}
