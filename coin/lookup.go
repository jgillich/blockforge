package coin

import (
	"strings"

	"gitlab.com/blockforge/blockforge/algo"
	"gitlab.com/blockforge/blockforge/algo/cryptonight"
	"gitlab.com/blockforge/blockforge/algo/ethash"
	"gitlab.com/blockforge/blockforge/stratum"
)

var Coins = []Coin{
	Coin{"Monero", "XMR", &cryptonight.Algo{}, stratum.ProtocolCryptonight},
	Coin{"Electroneum", "ETN", &cryptonight.Algo{}, stratum.ProtocolCryptonight},
	Coin{"IntenseCoin", "ITNS", &cryptonight.Algo{}, stratum.ProtocolCryptonight},
	Coin{"Sumokoin", "SUMO", &cryptonight.Algo{}, stratum.ProtocolCryptonight},
	Coin{"Bytecoin", "BCN", &cryptonight.Algo{}, stratum.ProtocolCryptonight},
	Coin{"Aeon", "AEON", &cryptonight.Algo{Lite: true}, stratum.ProtocolCryptonight},
	Coin{"Ethereum", "ETH", &ethash.Algo{}, stratum.ProtocolEthereum},
	Coin{"Ethereum Classic", "ETC", &ethash.Algo{}, stratum.ProtocolEthereum},
}

type Coin struct {
	LongName  string           `json:"long_name"`
	ShortName string           `json:"short_name"`
	Algo      algo.Algo        `json:"algo"`
	Protocol  stratum.Protocol `json:"protocol"`
}

func Lookup(coinName string) *Coin {
	coinName = strings.ToLower(coinName)
	for _, coin := range Coins {
		if strings.ToLower(coin.ShortName) == coinName {
			return &coin
		}
		if strings.ToLower(coin.LongName) == coinName {
			return &coin
		}
	}
	return nil
}
