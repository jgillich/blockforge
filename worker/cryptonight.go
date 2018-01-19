package worker

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"

	"gitlab.com/jgillich/autominer/hash"

	"gitlab.com/jgillich/autominer/stratum"
)

// NonceIndex is the starting location of nonce in blob
var NonceIndex = 78

// NonceWidth is the char width of nonce in blob
var NonceWidth = 8

func init() {
	for _, c := range []string{"xmr", "etn", "itns", "sumo"} {
		workers[c] = func(config Config) Worker {
			return NewCryptonight(config, false)
		}
	}
	for _, c := range []string{"aeon"} {
		workers[c] = func(config Config) Worker {
			return NewCryptonight(config, true)
		}
	}
}

type cryptonight struct {
	stratum stratum.Client
	light   bool
}

func NewCryptonight(config Config, light bool) Worker {
	return &cryptonight{
		stratum: config.Stratum,
		light:   light,
	}
}

func (w *cryptonight) Work() error {

	for {
		job := <-w.stratum.Jobs()

		log.Printf("working on new job '%v'", job.JobId)

		blob := []byte(job.Blob)

		target := math.MaxUint64 / uint64(math.MaxUint32/hexUint64LE([]byte(job.Target)))

		for nonce := hexUint32(blob[NonceIndex : NonceIndex+NonceWidth]); nonce < math.MaxUint32; nonce++ {
			blob := fmt.Sprintf("%v%v%v", job.Blob[:NonceIndex], fmtNonce(nonce), job.Blob[NonceIndex+NonceWidth:])

			input, err := hex.DecodeString(blob)
			if err != nil {
				return err
			}

			hash := hash.Cryptonight(input)

			val := hexUint64LE([]byte(hex.EncodeToString(hash)[48:]))

			if val < target {
				share := stratum.Share{
					MinerId: job.MinerId,
					JobId:   job.JobId,
					Result:  fmt.Sprintf("%x", hash),
					Nonce:   fmtNonce(nonce),
				}

				w.stratum.SubmitShare(&share)
			}

			if len(w.stratum.Jobs()) > 0 {
				break
			}
		}

	}
}

func (e *cryptonight) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: false,
		CUDA:   false,
	}
}
