package hash

// #include <stdint.h>
// #include "hash.h"
// #cgo LDFLAGS: -L${SRCDIR}/build/ -lhash
import "C"
import "unsafe"

func Cryptonight(data []byte) []byte {
	hash := make([]byte, 32)
	C.cryptonight((*C.char)(unsafe.Pointer(&data[0])), (C.uint32_t)(len(data)), (*C.char)(unsafe.Pointer(&hash[0])))
	return hash
}
