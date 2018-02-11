package cryptonight

import (
	"encoding/binary"
	"math"
	"sync/atomic"
	"time"

	"gitlab.com/blockforge/blockforge/hash"
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

func (work *Work) verify(lite bool, input []byte, nonce uint32) (ok bool, result []byte) {
	binary.LittleEndian.PutUint32(input[NonceIndex:], nonce)
	if lite {
		result = hash.CryptonightLite(input)
	} else {
		result = hash.Cryptonight(input)
	}
	ok = binary.LittleEndian.Uint64(result[24:]) < work.Target
	return
}

func (work *Work) Verify(lite bool, nonce uint32) (bool, []byte) {
	input := make([]byte, len(work.Input))
	copy(input, work.Input)
	return work.verify(lite, input, nonce)
}

func (work *Work) VerifySend(lite bool, nonce uint32, results chan<- Share) bool {
	if ok, result := work.Verify(lite, nonce); ok {
		results <- Share{
			JobId:  work.JobId,
			Result: result,
			Nonce:  nonce,
		}
		return true
	}
	return false
}

func (work *Work) VerifyRange(lite bool, size uint32, results chan<- Share) {
	input := make([]byte, len(work.Input))
	copy(input, work.Input)
	start := work.NextNonce(size)

	for i := start; i < start+size; i++ {
		ok, result := work.verify(lite, input, i)
		if ok {
			results <- Share{
				JobId:  work.JobId,
				Result: result,
				Nonce:  i,
			}
		}
	}
	return
}
