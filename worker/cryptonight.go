package worker

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"gitlab.com/jgillich/autominer/hash"

	"gitlab.com/jgillich/autominer/stratum"
)

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
	stratum *stratum.Client
	light   bool
}

func NewCryptonight(config Config, light bool) Worker {
	return &cryptonight{
		stratum: config.Stratum,
		light:   light,
	}
}

func (w *cryptonight) Work() error {
	job := <-w.stratum.Jobs

	for {
		log.Printf("working on new job '%v'", job.JobId)
		target, err := strconv.ParseUint(fmt.Sprintf("0x%v", job.Target), 0, 64)
		if err != nil {
			return err
		}

		for i := uint(0); i < math.MaxUint32; i++ {
			nonce := fmt.Sprintf("%x", i)
			input := fmt.Sprintf("%v%v%v", job.Blob[78:], nonce, job.Blob[:86])

			result := hash.Cryptonight([]byte(input))
			val, err := strconv.ParseUint("0x"+fmt.Sprintf("%x", result)[48:], 0, 64)
			if err != nil {
				return err
			}

			if val < target {
				share := stratum.Share{
					MinerId: job.MinerId,
					JobId:   job.JobId,
					Result:  fmt.Sprintf("%x", result),
					Nonce:   nonce,
				}

				w.stratum.SubmitShare(&share)
				log.Printf("found share %+v", share)
				continue
			}

			if len(w.stratum.Jobs) > 0 {
				job = <-w.stratum.Jobs
				continue
			}
		}

		return fmt.Errorf("nounce space exhausted")
	}
}

func (e *cryptonight) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: false,
		CUDA:   false,
	}
}
