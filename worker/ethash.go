package worker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"gitlab.com/blockforge/blockforge/hash"
	"gitlab.com/blockforge/blockforge/stratum"
)

func init() {
	for _, c := range []string{"ETH", "ETC"} {
		workers[c] = newEthash
	}
}

var maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

type ethash struct {
	config Config
}

func newEthash(config Config) Worker {
	return &ethash{config}
}

func (w *ethash) Work() error {
	for {
		j := w.config.Stratum.GetJob()
		if j == nil {
			return nil
		}

		job := j.(stratum.NicehashJob)

		//target := new(big.Int).Div(maxUint256, job.Difficulty)
		fmt.Printf("diff: %v\n", job.Difficulty)

		//target := new(big.Int).Div(maxUint256, job.Difficulty)

		fmt.Printf("target: %x\n", diffToTarget(job.Difficulty))

		seedHash, err := hex.DecodeString(strings.TrimPrefix(job.SeedHash, "0x"))
		if err != nil {
			return err
		}

		_, err = hash.NewEthash(seedHash)
		if err != nil {
			return err
		}

	}
}

// diffToTarget converts a stratum pool difficulty to target
func diffToTarget(diff float32) *big.Int {
	var k int

	for k = 6; k > 0 && diff > 1.0; k-- {
		diff /= 4294967296.0
	}

	m := uint64(4294901760.0 / diff)

	t := make([]byte, 32)

	if m == 0 && k == 6 {
		for i := 0; i < 32; i++ {
			t[i] = 0xFF
		}
	} else {
		binary.LittleEndian.PutUint64(t[k*4:], m)
	}

	reverse := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reverse[31-i] = t[i]
	}
	return new(big.Int).SetBytes(reverse)
}

func (w *ethash) Stats() Stats {
	stats := Stats{
		CPUStats: []CPUStats{},
		GPUStats: []GPUStats{},
	}

	return stats
}

func (w *ethash) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: false,
		CUDA:   false,
	}
}
