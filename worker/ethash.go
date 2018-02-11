package worker

import (
	crand "crypto/rand"
	"encoding/hex"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"time"

	"gitlab.com/blockforge/blockforge/algo/ethash"
	"gitlab.com/blockforge/blockforge/hash"
	"gitlab.com/blockforge/blockforge/log"
)

var maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

type Ethash struct {
	Work   <-chan *ethash.Work
	Shares chan<- ethash.Share

	config Config
	// random source for nonces
	rand *rand.Rand

	hash     *hash.Ethash
	seedhash string

	lock sync.RWMutex
}

func (worker *Ethash) Configure(config Config) error {
	seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return err
	}

	worker.config = config
	worker.rand = rand.New(rand.NewSource(seed.Int64()))

	return nil
}

func (worker *Ethash) Start() error {
	totalThreads := 1

	workChannels := make([]chan *ethash.Work, totalThreads)
	for i := 0; i < totalThreads; i++ {
		workChannels[i] = make(chan *ethash.Work, 1)
		index := i
		defer close(workChannels[index])
	}

	go worker.thread(workChannels[0])

	for work := range worker.Work {
		if worker.seedhash != work.Seedhash {
			worker.seedhash = work.Seedhash
			seedhash, err := hex.DecodeString(strings.TrimPrefix(work.Seedhash, "0x"))
			if err != nil {
				return err
			}

			log.Info("DAG is being initialized, this may take a while")
			worker.lock.Lock()
			worker.hash, err = hash.NewEthash(seedhash)
			worker.lock.Unlock()
			if err != nil {
				return err
			}
			log.Info("DAG initialized")
		}

		for _, ch := range workChannels {
			ch <- work
		}
	}

	return nil
}

func (worker *Ethash) thread(workChan chan *ethash.Work) {
	work := <-workChan
	var ok bool
	start := time.Now()
	hashes := float64(0)

	for {
		select {
		case work, ok = <-workChan:
			if !ok {
				return
			}

			log.Infof("ethash hashrate %v H/s", hashes/time.Since(start).Seconds())

			start = time.Now()
			hashes = 0
		default:
			worker.lock.RLock()
			if _, err := work.VerifySend(worker.hash, work.ExtraNonce+work.Nonce, worker.Shares); err != nil {
				log.Error(err)
			}
			worker.lock.RUnlock()
			work.Nonce++
			hashes++
		}
	}
}

func (w *Ethash) Stats() Stats {
	stats := Stats{
		CPUStats: []CPUStats{},
		GPUStats: []GPUStats{},
	}

	return stats
}

func (w *Ethash) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: false,
		CUDA:   false,
	}
}
