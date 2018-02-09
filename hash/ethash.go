package hash

// #include <stdint.h>
// #include "ethash/ethash.h"
// #cgo LDFLAGS: -L${SRCDIR}/build/ -lhash
import "C"
import (
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
)

type Ethash struct {
	light C.ethash_light_t
	full  C.ethash_full_t
}

func NewEthash(blockNumber uint64) *Ethash {
	light := C.ethash_light_new(C.uint64_t(blockNumber))
	full := C.ethash_full_new(light, nil)
	return &Ethash{light, full}
}

func (e *Ethash) Compute(hash common.Hash, nonce uint64) (success bool, mixDigest, result common.Hash) {
	ret := C.ethash_full_compute(e.full, hashToH256(hash), C.uint64_t(nonce))
	return bool(ret.success), h256ToHash(ret.mix_hash), h256ToHash(ret.result)
}

func (e *Ethash) Release() {
	C.ethash_light_delete(e.light)
	C.ethash_full_delete(e.full)
}

func h256ToHash(in C.ethash_h256_t) common.Hash {
	return *(*common.Hash)(unsafe.Pointer(&in.b))
}

func hashToH256(in common.Hash) C.ethash_h256_t {
	return C.ethash_h256_t{b: *(*[32]C.uint8_t)(unsafe.Pointer(&in[0]))}
}
