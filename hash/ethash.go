package hash

// #include <stdint.h>
// #include "ethash/ethash.h"
// #cgo LDFLAGS: -L${SRCDIR}/build/ -lhash
import "C"
import "math/big"

var maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

type Ethash struct {
	light C.ethash_light_t
	full  C.ethash_full_t
}

func NewEthash(blockNumber uint64) *Ethash {
	light := C.ethash_light_new(C.uint64_t(blockNumber))
	full := C.ethash_full_new(light, nil)
	return &Ethash{light, full}
}

func (e *Ethash) Compute(header []uint8, nonce uint64) (bool, *[32]uint8, *[32]uint8) {
	// TODO can we avoid copying here?
	b := [32]C.uint8_t{}
	for i := 0; i < 32; i++ {
		b[i] = C.uint8_t(header[i])
	}
	val := C.ethash_full_compute(e.full, C.ethash_h256_t{b}, C.uint64_t(nonce))

	if !val.success {
		return false, nil, nil
	}

	result := [32]uint8{}
	for i := 0; i < 32; i++ {
		result[i] = uint8(val.result.b[i])
	}
	mixHash := [32]uint8{}
	for i := 0; i < 32; i++ {
		mixHash[i] = uint8(val.mix_hash.b[i])
	}

	return true, &result, &mixHash
}

func (e *Ethash) Release() {
	C.ethash_light_delete(e.light)
	C.ethash_full_delete(e.full)
}
