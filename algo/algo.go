package algo

type Algo interface {
	MarshalJSON() ([]byte, error)
}

/*
var (
	Cryptonight     Algo = "cryptonight"
	CryptonightLite Algo = "cryptonight-lite"
	Ethash          Algo = "ethash"
)
*/
