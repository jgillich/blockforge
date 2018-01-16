package hash

import (
	"encoding/hex"
	"testing"
)

func TestCryptonote(t *testing.T) {
	blob, _ := hex.DecodeString("01009091e4aa05ff5fe4801727ed0c1b8b339e1a0054d75568fec6ba9c4346e88b10d59edbf6858b2b00008a63b2865b65b84d28bb31feb057b16a21e2eda4bf6cc6377e3310af04debe4a01")
	hashBytes := Cryptonight(blob)
	hash := hex.EncodeToString(hashBytes)
	if hash != "a70a96f64a266f0f59e4f67c4a92f24fe8237c1349f377fd2720c9e1f2970400" {
		t.Error("Invalid hash")
	}
}
