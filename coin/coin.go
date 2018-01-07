package coin

var Coins = map[string]Coin{}

type Coin interface {
	Miner(MinerConfig) (Miner, error)
	//Info() CoinInfo
}

/*
type CoinInfo struct {
	SupportsCPU    bool
	SupportsOpenGL bool
	SupportsCUDA   bool
}
*/
