package worker

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"time"

	"gitlab.com/blockforge/blockforge/hash"
	"gitlab.com/blockforge/blockforge/log"
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
	// random source for nonces
	rand *rand.Rand

	hash     *hash.Ethash
	seedhash string

	lock sync.RWMutex
}

type ethashWork struct {
	jobId      string
	header     []byte
	target     *big.Int
	extraNonce uint64
	nonce      uint64
}

func newEthash(config Config) Worker {
	seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}

	return &ethash{
		config: config,
		rand:   rand.New(rand.NewSource(seed.Int64())),
	}
}

func (w *ethash) Work() error {
	totalThreads := 1

	workChannels := make([]chan *ethashWork, totalThreads)
	for i := 0; i < totalThreads; i++ {
		workChannels[i] = make(chan *ethashWork, 1)
		index := i
		defer close(workChannels[index])
	}

	go w.thread(workChannels[0])

	for {
		j := w.config.Stratum.GetJob()
		if j == nil {
			return nil
		}

		job := j.(stratum.NicehashJob)
		target := diffToTarget(job.Difficulty)

		log.Debugf("ethash difficulty '%v', target '%x'", job.Difficulty, target)

		if w.seedhash != job.SeedHash {
			w.seedhash = job.SeedHash
			seedHash, err := hex.DecodeString(strings.TrimPrefix(job.SeedHash, "0x"))
			if err != nil {
				return err
			}

			log.Info("initializing DAG, this may take a while")
			w.lock.Lock()
			w.hash, err = hash.NewEthash(seedHash)
			w.lock.Unlock()
			if err != nil {
				return err
			}
			log.Info("DAG initialized")
		}

		header, err := hex.DecodeString(strings.TrimPrefix(job.SeedHash, "0x"))
		if err != nil {
			return err
		}

		for i := len(job.ExtraNonce); i < 16; i++ {
			job.ExtraNonce += "0"
		}
		extraNonceBytes, err := hex.DecodeString(job.ExtraNonce)
		if err != nil {
			return err
		}
		extraNonce := binary.BigEndian.Uint64(extraNonceBytes)

		for _, ch := range workChannels {
			ch <- &ethashWork{
				jobId:      job.JobId,
				header:     header,
				target:     target,
				extraNonce: extraNonce,
				// TODO ?
				nonce: uint64(w.rand.Uint32()),
			}
		}
	}
}

func (w *ethash) thread(workChan chan *ethashWork) {
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
			w.lock.RLock()
			success, _, result := w.hash.Compute(work.header, work.extraNonce+work.nonce)
			w.lock.RUnlock()
			if !success {
				log.Error("ethash compute failed")
			} else {
				if new(big.Int).SetBytes(result[:]).Cmp(work.target) <= 0 {
					w.config.Stratum.SubmitShare(stratum.NicehashShare{
						JobId: work.jobId,
						Nonce: fmt.Sprintf("%x", work.nonce),
					})
				}
			}

			work.nonce++
			hashes++
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
