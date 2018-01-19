package hash

import (
	"encoding/hex"
	"testing"
)

func TestCryptonote(t *testing.T) {
	blob, _ := hex.DecodeString("0606fcfc85d305d5f238078d3eaf897b43bc2548024c4c753c15584cd30a9323be296e1554ecb50000001bd9286f6d087f92749e8f094090af879f8bc16ddce6db71e5ed745e7ed806a98e09")
	hashBytes := Cryptonight(blob)
	hash := hex.EncodeToString(hashBytes)
	if hash != "03deef54ac208e5c4c41b608fa3c37436c5350858766d332fffbd8b06efc0700" {
		t.Error("Invalid hash")
	}
}
