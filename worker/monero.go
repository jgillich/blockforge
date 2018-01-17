package worker

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"gitlab.com/jgillich/autominer/hash"

	"gitlab.com/jgillich/autominer/stratum"
)

type MoneroWorker struct {
	stratum *stratum.Client
}

func NewMoneroWorker(stratum *stratum.Client) MoneroWorker {
	return MoneroWorker{stratum}
}

func (w *MoneroWorker) Work() error {

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
