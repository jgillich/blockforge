package worker

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"time"

	metrics "github.com/armon/go-metrics"
	"gitlab.com/blockforge/blockforge/algo/ethash"
	"gitlab.com/blockforge/blockforge/log"
)

var maxUint256 = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))

type Ethash struct {
	Work   <-chan *ethash.Work
	Shares chan<- ethash.Share

	config Config
	// random source for nonces
	rand *rand.Rand

	hash     *ethash.Ethash
	seedhash string
	lock     sync.RWMutex

	metrics *metrics.Metrics
}

func (worker *Ethash) Configure(config Config) error {
	seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return err
	}

	worker.config = config
	worker.rand = rand.New(rand.NewSource(seed.Int64()))

	worker.metrics = config.Metrics

	return nil
}

func (worker *Ethash) Start() error {
	totalThreads := len(worker.config.CLDevices)
	for _, c := range worker.config.Processors {
		totalThreads += c.Threads
	}

	workChannels := make([]chan *ethash.Work, totalThreads)
	for i := 0; i < totalThreads; i++ {
		workChannels[i] = make(chan *ethash.Work, 1)
		index := i
		defer close(workChannels[index])
	}

	for work := range worker.Work {
		if worker.seedhash != work.Seedhash {
			worker.seedhash = work.Seedhash
			seedhash, err := hex.DecodeString(strings.TrimPrefix(work.Seedhash, "0x"))
			if err != nil {
				return err
			}

			// when DAG changes, we shutdown and recreate all threads
			for i := 0; i < totalThreads; i++ {
				close(workChannels[i])
				workChannels[i] = make(chan *ethash.Work, 1)
			}

			log.Info("DAG is being initialized, this may take a while")
			worker.lock.Lock()
			worker.hash, err = ethash.NewEthash(seedhash)
			worker.lock.Unlock()
			if err != nil {
				return err
			}
			log.Info("DAG initialized")

			for cpuIndex, conf := range worker.config.Processors {
				for i := 0; i < conf.Threads; i++ {
					key := []string{"cpu", fmt.Sprintf("%v", cpuIndex), fmt.Sprintf("%v", i)}
					go worker.thread(key, workChannels[len(worker.config.CLDevices)+i])
				}
			}

			if len(worker.config.CLDevices) > 0 {
				for i, d := range worker.config.CLDevices {
					cl, err := newEthashCL(d, worker.hash)
					if err != nil {
						return err
					}
					key := []string{"opencl", fmt.Sprintf("%v", d.Device.Platform.Index), fmt.Sprintf("%v", d.Device.Index)}

					go worker.clThread(key, cl, workChannels[i])
				}
			}
		}

		for _, ch := range workChannels {
			ch <- work
		}
	}

	return nil
}

func (worker *Ethash) thread(key []string, workChan chan *ethash.Work) {
	work := <-workChan
	var ok bool

	for {
		select {
		case work, ok = <-workChan:
			if !ok {
				return
			}

		default:
			start := time.Now()
			worker.lock.RLock()
			if err := work.VerifyRange(worker.hash, uint64(worker.rand.Uint32()), 10*1024, worker.Shares); err != nil {
				workerError(err)
			}
			worker.lock.RUnlock()
			worker.metrics.IncrCounter(key, float32(10*1024/time.Since(start).Seconds()))
		}
	}
}

func (worker *Ethash) clThread(key []string, cl *ethashCL, workChan chan *ethash.Work) {
	defer cl.Release()

	work := <-workChan
	if err := cl.Update(work.Header, work.Target); err != nil {
		workerError(err)
	}

	var ok bool
	var results [2]uint32

	for {
		select {
		case work, ok = <-workChan:
			if !ok {
				return
			}
			if err := cl.Update(work.Header, work.Target); err != nil {
				workerError(err)
			}

		default:
			start := time.Now()
			worker.lock.RLock()
			startNonce := uint64(worker.rand.Uint32())
			if err := cl.Run(work.ExtraNonce+startNonce, results); err != nil {
				workerError(err)
			}
			if results[0] > 0 {
				worker.Shares <- ethash.Share{
					JobId: work.JobId,
					Nonce: startNonce + uint64(results[1]),
				}
			}
			worker.lock.RUnlock()
			worker.metrics.IncrCounter(key, float32(10*1024/time.Since(start).Seconds()))
		}
	}

}

func (w *Ethash) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: true,
		CUDA:   false,
	}
}
