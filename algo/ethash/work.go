package ethash

// #include <stdint.h>
// #include "../../hash/ethash/ethash.h"
// #cgo LDFLAGS: -L${SRCDIR}/../../hash/build/ -lhash
import "C"

import (
	"encoding/binary"
	"fmt"
	"math/big"
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

	// reverse
	for i := 0; i < len(t)/2; i++ {
		j := len(t) - i - 1
		t[i], t[j] = t[j], t[i]
	}

	return new(big.Int).SetBytes(t)
}

func (work *Work) Verify(hash *Ethash, nonce uint64) (bool, error) {
	ret := C.ethash_full_compute(hash.full, hashToH256(work.Header), C.uint64_t(nonce))
	success, result := bool(ret.success), h256ToHash(ret.result)
	if !success {
		return false, fmt.Errorf("ethash compute failed")
	} else if new(big.Int).SetBytes(result[:]).Cmp(work.Target) <= 0 {
		return true, nil
	}
	return false, nil
}

func (work *Work) VerifySend(hash *Ethash, nonce uint64, results chan<- Share) (bool, error) {
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

func (work *Work) VerifyRange(hash *Ethash, start uint64, size uint64, results chan<- Share) error {
	end := start + size + work.ExtraNonce
	for i := start + work.ExtraNonce; i < end; i++ {
		if _, err := work.VerifySend(hash, i, results); err != nil {
			return err
		}
	}
	return nil
}
