package cryptonight

var (
	CryptonightMemory     uint32 = 2097152
	CryptonightMask       uint32 = 0x1FFFF0
	CryptonightIter       uint32 = 0x80000
	CryptonightLiteMemory uint32 = 1048576
	CryptonightLiteMask   uint32 = 0xFFFF0
	CryptonightLiteIter   uint32 = 0x40000
)

// NonceIndex is the starting location of nonce in binary blob
var NonceIndex = 39
