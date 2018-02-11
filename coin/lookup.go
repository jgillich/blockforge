package coin

import (
	"strings"

	"gitlab.com/blockforge/blockforge/algo"
	"gitlab.com/blockforge/blockforge/stratum"
)

var Coins = []Coin{
	Coin{"Monero", "XMR", algo.Cryptonight, stratum.ProtocolCryptonight},
	Coin{"Electroneum", "ETN", algo.Cryptonight, stratum.ProtocolCryptonight},
	Coin{"IntenseCoin", "ITNS", algo.Cryptonight, stratum.ProtocolCryptonight},
	Coin{"Sumokoin", "SUMO", algo.Cryptonight, stratum.ProtocolCryptonight},
	Coin{"Bytecoin", "BCN", algo.Cryptonight, stratum.ProtocolCryptonight},
	Coin{"Aeon", "AEON", algo.CryptonightLite, stratum.ProtocolCryptonight},
	Coin{"Ethereum", "ETH", algo.Ethash, stratum.ProtocolEthereum},
	Coin{"Ethereum Classic", "ETC", algo.Ethash, stratum.ProtocolEthereum},
}

type Coin struct {
	LongName  string
	ShortName string
	Algo      algo.Algo
	Protocol  stratum.Protocol
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
