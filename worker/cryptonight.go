package worker

import (
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"gitlab.com/jgillich/autominer/hash"
	"gitlab.com/jgillich/autominer/log"

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
	config  Config
	stratum stratum.Client
	light   bool
}

func NewCryptonight(config Config, light bool) Worker {
	return &cryptonight{
		stratum: config.Stratum,
		config:  config,
		light:   light,
	}
}

func (w *cryptonight) Work() error {
	closer := make(chan int, 0)

	for {
		job, ok := <-w.stratum.Jobs()

		close(closer)
		closer = make(chan int, 0)

		if !ok {
			return nil
		}

		numThreads := 0

		for _, cpu := range w.config.CPUSet {
			numThreads += cpu.Threads
		}

		nounceStepping := uint32(math.MaxUint32 / numThreads)
		nonce := uint32(0)

		for _, cpu := range w.config.CPUSet {
			for i := 0; i < cpu.Threads; i++ {
				log.Debugf("starting thread for job '%v'", job.JobId)
				log.Debugf("nonce start '%v' end '%v'", nonce, nonce+nounceStepping)
				go w.cpuThread(job, nonce, nonce+nounceStepping, closer)
				nonce += nounceStepping
			}
		}

	}
}

func (w *cryptonight) cpuThread(job stratum.Job, nonceStart uint32, nonceEnd uint32, closer chan int) {
	target := math.MaxUint64 / uint64(math.MaxUint32/hexUint64LE([]byte(job.Target)))
	hashes := float64(0)
	startTime := time.Now()

	defer func() {
		log.Infof("thread hashes: %.0f H/s", hashes/time.Since(startTime).Seconds())
	}()

	for nonce := nonceStart; nonce < nonceEnd; nonce++ {
		select {
		default:
			blob := fmt.Sprintf("%v%v%v", job.Blob[:NonceIndex], fmtNonce(nonce), job.Blob[NonceIndex+NonceWidth:])

			input, err := hex.DecodeString(blob)
			if err != nil {
				log.Errorf("malformed blob: '%v'", blob)
				return
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
			hashes++
		case <-closer:
			log.Debugf("stopping thread for job '%v'", job.JobId)
			return
		}
	}
}

func (w *cryptonight) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: false,
		CUDA:   false,
	}
}
