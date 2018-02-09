package hash

// #include <stdint.h>
// #include "ethash/ethash.h"
// #cgo LDFLAGS: -L${SRCDIR}/build/ -lhash
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/ethereum/go-ethereum/crypto/sha3"
)

var EthashEpochLength uint64 = 30000

type Ethash struct {
	light C.ethash_light_t
	full  C.ethash_full_t
}

func NewEthash(seedHash []byte) (*Ethash, error) {
	blockNumber, err := seedHashToBlockNum(seedHash)
	if err != nil {
		return nil, err
	}

	light := C.ethash_light_new(C.uint64_t(blockNumber))
	full := C.ethash_full_new(light, nil)

	return &Ethash{light, full}, nil
}

func (e *Ethash) Compute(hash []byte, nonce uint64) (success bool, mixDigest, result [32]byte) {
	ret := C.ethash_full_compute(e.full, hashToH256(hash), C.uint64_t(nonce))
	return bool(ret.success), h256ToHash(ret.mix_hash), h256ToHash(ret.result)
}

func (e *Ethash) Release() {
	C.ethash_light_delete(e.light)
	C.ethash_full_delete(e.full)
}

func seedHashToBlockNum(seedHash []byte) (uint64, error) {
	data := make([]byte, 32)
	hash := sha3.NewKeccak256()
	var epoch uint64

	for epoch = 0; epoch < 2048; epoch++ {
		hash.Reset()
		_, err := hash.Write(data)
		if err != nil {
			return 0, err
		}
		data = hash.Sum(nil)
		equal := true
		for i := 0; i < 32; i++ {
			if seedHash[i] != data[i] {
				equal = false
				break
			}
		}
		if equal {
			return epoch * EthashEpochLength, nil
		}
	}

	return 0, fmt.Errorf("no block number found for seed hash '%x'", seedHash)
}

func h256ToHash(in C.ethash_h256_t) [32]byte {
	return *(*[32]byte)(unsafe.Pointer(&in.b))
}

func hashToH256(in []byte) C.ethash_h256_t {
	return C.ethash_h256_t{b: *(*[32]C.uint8_t)(unsafe.Pointer(&in[0]))}
}
