package cryptonight

import "gitlab.com/jgillich/autominer/coin"

type Miner struct {
	coin  string
	light bool
}

func NewMiner(config coin.MinerConfig, light bool) (coin.Miner, error) {
	miner := Miner{coin: config.Coin, light: light}

	return &miner, nil
}

func (m *Miner) Start() error {
	//xmrstak.ExecutorInst().Ex_start(true)

	return nil
}

func (m *Miner) Stop() {

}

func (m *Miner) Stats() coin.MinerStats {
	hashrate := 0
	return coin.MinerStats{
		Coin:     m.coin,
		Hashrate: float32(hashrate),
	}
}
