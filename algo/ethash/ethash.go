package ethash

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"gitlab.com/blockforge/blockforge/hash"
)

type Work struct {
	JobId      string
	Seedhash   string
	Header     []byte
	Target     *big.Int
	ExtraNonce uint64
	Nonce      uint64
}

// DiffToTarget converts a stratum pool difficulty to target
func DiffToTarget(diff float32) *big.Int {
	var k int

	for k = 6; k > 0 && diff > 1.0; k-- {
		diff /= 4294967296.0
	}

	m := uint64(4294901760.0 / diff)

	t := make([]byte, 32)

	if m == 0 && k == 6 {
		for i := 0; i < 32; i++ {
			t[i] = 0xFF
		}
	} else {
		binary.LittleEndian.PutUint64(t[k*4:], m)
	}

	reverse := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reverse[31-i] = t[i]
	}
	return new(big.Int).SetBytes(reverse)
}

func (work *Work) Verify(hash *hash.Ethash, nonce uint64) (bool, error) {
	success, _, result := hash.Compute(work.Header, nonce)
	if !success {
		return false, fmt.Errorf("ethash compute failed")
	} else if new(big.Int).SetBytes(result[:]).Cmp(work.Target) <= 0 {
		return true, nil
	}
	return false, nil
}

func (work *Work) VerifySend(hash *hash.Ethash, nonce uint64, results chan<- Share) (bool, error) {
	if ok, err := work.Verify(hash, nonce); ok {
		results <- Share{
			JobId: work.JobId,
			Nonce: nonce,
		}
		return true, err
	} else {
		return false, err
	}
}

func (work *Work) VerifyRange(hash *hash.Ethash, start uint64, size uint64, results chan<- Share) error {
	for i := start + work.ExtraNonce; i < start+size+work.ExtraNonce; i++ {
		if _, err := work.VerifySend(hash, i, results); err != nil {
			return err
		}
	}
	return nil
}
