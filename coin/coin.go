package coin

var Coins = map[string]Coin{}

type Coin interface {
	Miner(MinerConfig) (Miner, error)
	Info() Info
}

type Info struct {
	SupportsCPU    bool
	SupportsOpenCL bool
	SupportsCUDA   bool
}
