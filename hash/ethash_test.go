package hash

import (
	"encoding/hex"
	"testing"
)

func TestEthash(t *testing.T) {
	/*
			Height =     22,
		                    HashNoNonce ="372eca2454ead349c3df0ab5d00b0b706b23e49d469387db91811cee0358fc6d".HexToByteArray(),
		                    Difficulty = new BigInteger(132416),
		                    Nonce =      0x495732e0ed7a801c,
		MixDigest = "2f74cdeb198af0b9abe65d22d372e22fb2d474371774a9583c1cc427a07939f5".HexToByteArray(),
	*/

	ethash := NewEthash(22)

	header, err := hex.DecodeString("372eca2454ead349c3df0ab5d00b0b706b23e49d469387db91811cee0358fc6d")
	if err != nil {
		t.Fatal(err)
	}

	success, result, mixHash := ethash.Compute(header, 132416)
	if !success {
		t.Fatal("compute failed")
	}

	t.Logf("result: %v", result)
	t.Logf("mixHash: %v", mixHash)
	// TODO ...

}
