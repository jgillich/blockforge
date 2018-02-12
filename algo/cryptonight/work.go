package cryptonight

// #include <stdint.h>
// #include "../../hash/cryptonight/hash-ops.h"
// #cgo LDFLAGS: -L${SRCDIR}/../../hash/build/ -lhash
import "C"

import (
	"encoding/binary"
	"math"
	"sync/atomic"
	"time"
	"unsafe"

	"gitlab.com/blockforge/blockforge/log"
)

type Work struct {
	JobId  string
	Input  []byte
	Target uint64
	nonce  uint32
}

func (work *Work) NextNonce(size uint32) uint32 {
	for {
		val := atomic.LoadUint32(&work.nonce)
		if val > math.MaxUint32-size {
			log.Error("nonce space exceeded")
			time.Sleep(time.Second * 5)
			return val
		}
		if atomic.CompareAndSwapUint32(&work.nonce, val, val+size) {
			return val
		}
	}
}

func (work *Work) verify(lite bool, input []byte, nonce uint32, result *[32]byte) bool {
	binary.LittleEndian.PutUint32(input[NonceIndex:], nonce)
	if lite {
		C.cn_slow_hash_lite(unsafe.Pointer(&input[0]), (C.size_t)(len(input)), (*C.char)(unsafe.Pointer(&result[0])))
	} else {
		C.cn_slow_hash(unsafe.Pointer(&input[0]), (C.size_t)(len(input)), (*C.char)(unsafe.Pointer(&result[0])))
	}
	return binary.LittleEndian.Uint64(result[24:]) < work.Target
}

func (work *Work) Verify(lite bool, nonce uint32, result *[32]byte) bool {
	input := make([]byte, len(work.Input))
	copy(input, work.Input)
	return work.verify(lite, input, nonce, result)
}

func (work *Work) VerifySend(lite bool, nonce uint32, results chan<- Share) bool {
	var result [32]byte
	if work.Verify(lite, nonce, &result) {
		results <- Share{
			JobId:  work.JobId,
			Result: result[:],
			Nonce:  nonce,
		}
		return true
	}
	return false
}

func (work *Work) VerifyRange(lite bool, size uint32, results chan<- Share) {
	var result [32]byte
	input := make([]byte, len(work.Input))
	copy(input, work.Input)
	start := work.NextNonce(size)

	for i := start; i < start+size; i++ {
		ok := work.verify(lite, input, i, &result)
		if ok {
			results <- Share{
				JobId:  work.JobId,
				Result: result[:],
				Nonce:  i,
			}
		}
	}
}
