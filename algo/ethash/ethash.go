package ethash

// #include <stdint.h>
// #include "../../hash/ethash/internal.h"
// #include "../../hash/ethash/io.h"
// #cgo LDFLAGS: -L${SRCDIR}/../../hash/build/ -lhash
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/ethereum/go-ethereum/crypto/sha3"
)

var EthashEpochLength uint64 = 30000

type Light struct {
	Cache    []byte
	DataSize int
	seed     C.ethash_h256_t
	light    C.ethash_light_t
}

type Full struct {
	Light *Light
	DAG   []byte
	full  C.ethash_full_t
}

func NewLight(seedhash []byte) (*Light, error) {
	blockNumber, err := seedHashToBlockNum(seedhash)
	if err != nil {
		return nil, err
	}

	seed := hashToH256(seedhash)

	light := C.ethash_light_new_internal(C.ethash_get_cachesize(C.uint64_t(blockNumber)), &seed)
	light.block_number = C.uint64_t(blockNumber)

	cache := C.GoBytes(unsafe.Pointer(light.cache), C.int(light.cache_size))

	datasize := int(C.ethash_get_datasize(light.block_number))

	return &Light{
		cache,
		datasize,
		seed,
		light,
	}, nil
}

func NewFull(light *Light) (*Full, error) {
	dir := make([]byte, 256)
	if !C.ethash_get_default_dirname((*C.char)(unsafe.Pointer(&dir[0])), 256) {
		return nil, fmt.Errorf("failed to determine ethash dag storage directory")
	}

	fullsize := C.ethash_get_datasize(light.light.block_number)
	full := C.ethash_full_new_internal((*C.char)(unsafe.Pointer(&dir[0])), light.seed, fullsize, light.light, nil)

	dag := C.GoBytes(unsafe.Pointer(C.ethash_full_dag(full)), C.int(C.ethash_full_dag_size(full)/4))

	return &Full{
		light,
		dag,
		full,
	}, nil
}

func (e *Light) Release() {
	C.ethash_light_delete(e.light)
}

func (e *Full) Release() {
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

type Algo struct {
}

func (algo *Algo) MarshalJSON() ([]byte, error) {
	return json.Marshal("ethash")
}
