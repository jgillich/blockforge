#include "cryptonight/hash-ops.h"

void cryptonight(const char *data, uint32_t length, char *hash) {
  cn_slow_hash(data, length, hash);
}
