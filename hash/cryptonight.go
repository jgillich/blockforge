package hash

// #include <stdint.h>
// #include "cryptonight/hash-ops.h"
// #cgo LDFLAGS: -L${SRCDIR}/build/ -lhash
import "C"
import "unsafe"

func Cryptonight(data []byte) []byte {
	hash := make([]byte, 32)
	C.cn_slow_hash(unsafe.Pointer(&data[0]), (C.size_t)(len(data)), (*C.char)(unsafe.Pointer(&hash[0])))
	return hash
}

func CryptonightLite(data []byte) []byte {
	hash := make([]byte, 32)
	C.cn_slow_hash_lite(unsafe.Pointer(&data[0]), (C.size_t)(len(data)), (*C.char)(unsafe.Pointer(&hash[0])))
	return hash
}
