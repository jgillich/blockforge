package worker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/blockforge/blockforge/hash"
	"gitlab.com/blockforge/blockforge/log"

	"gitlab.com/blockforge/blockforge/stratum"
)

var CryptonightMemory uint32 = 2097152
var CryptonightMask uint32 = 0x1FFFF0
var CryptonightIter uint32 = 0x80000

var CryptonightLiteMemory uint32 = 1048576
var CryptonightLiteMask uint32 = 0xFFFF0
var CryptonightLiteIter uint32 = 0x40000

// NonceIndex is the starting location of nonce in binary blob
var NonceIndex = 39

// NonceWidth is the char width of nonce in binary blob
var NonceWidth = 4

func init() {
	for _, c := range []string{"XMR", "ETN", "ITNS", "SUMO"} {
		workers[c] = func(config Config) Worker {
			return NewCryptonight(config, false)
		}
	}
	for _, c := range []string{"AEON"} {
		workers[c] = func(config Config) Worker {
			return NewCryptonight(config, true)
		}
	}
}

type cryptonight struct {
	clDevices  []CLDeviceConfig
	processors []ProcessorConfig
	stratum    stratum.Client
	lite       bool
	cpuStats   map[int]map[int]float32
	gpuStats   map[int]map[int]float32
	statMu     sync.RWMutex
}

type cryptonightWork struct {
	input  []byte
	target uint64
	nonce  uint32
}

type cryptonightShare struct {
	result []byte
	nonce  uint32
}

func (w *cryptonightWork) nextNonce(size uint32) uint32 {
	// TODO check for overflow
	for {
		val := atomic.LoadUint32(&w.nonce)
		if atomic.CompareAndSwapUint32(&w.nonce, val, val+size) {
			return val
		}
	}
}

func NewCryptonight(config Config, lite bool) Worker {
	return &cryptonight{
		clDevices:  config.CLDevices,
		processors: config.Processors,
		stratum:    config.Stratum,
		lite:       lite,
		cpuStats:   map[int]map[int]float32{},
		gpuStats:   map[int]map[int]float32{},
	}
}

func (w *cryptonight) Work() error {

	totalThreads := len(w.clDevices)
	for _, c := range w.processors {
		totalThreads += c.Threads
	}

	workChannels := make([]chan *cryptonightWork, totalThreads)
	for i := 0; i < totalThreads; i++ {
		workChannels[i] = make(chan *cryptonightWork, 1)
		defer close(workChannels[i])
	}

	shareChan := make(chan cryptonightShare, 10)
	defer close(shareChan)

	if len(w.clDevices) > 0 {
		for i, device := range w.clDevices {
			w.gpuStats[device.Device.Platform.Index] = map[int]float32{}
			worker, err := NewCryptonightCLWorker(device, w.lite)
			if err != nil {
				return err
			}
			go w.gpuThread(device.Device.Platform.Index, i, worker, workChannels[i], shareChan)
		}
	}

	for cpuIndex, conf := range w.processors {
		w.cpuStats[cpuIndex] = map[int]float32{}
		for i := 0; i < conf.Threads; i++ {
			go w.cpuThread(cpuIndex, i, workChannels[len(w.clDevices)+i], shareChan)
		}
	}

	var job stratum.Job

	go func() {
		for share := range shareChan {
			w.stratum.SubmitShare(&stratum.Share{
				MinerId: job.MinerId,
				JobId:   job.JobId,
				Result:  fmt.Sprintf("%x", share.result),
				Nonce:   fmt.Sprintf("%08x", share.nonce),
			})
		}
	}()

	for j := range w.stratum.Jobs() {
		job = j
		work, err := w.getWork(job)
		if err != nil {
			log.Errorw(err.Error(), "job", job)
			continue
		}
		for _, ch := range workChannels {
			ch <- work
		}
	}

	return nil
}

func (w *cryptonight) getWork(job stratum.Job) (*cryptonightWork, error) {
	input, err := hex.DecodeString(job.Blob)
	if err != nil {
		log.Errorw("malformed blob", "job", job)
		return nil, errors.New("malformed blob")
	}

	t, err := hex.DecodeString(job.Target)
	if err != nil {
		return nil, errors.New("malformed target")
	}

	var target uint64
	switch len(job.Target) {
	case 8:
		t32 := uint64(binary.LittleEndian.Uint32(t))
		target = math.MaxUint64 / (math.MaxUint32 / t32)
	case 16:
		target = binary.LittleEndian.Uint64(t)
	default:
		return nil, errors.New("unsupported target length")
	}

	log.Infof("job difficulty %v", math.MaxUint64/target)

	return &cryptonightWork{
		input:  input,
		target: target,
		nonce:  0,
	}, nil
}

func (w *cryptonight) gpuThread(platform, index int, cl *CryptonightCLWorker, workChan chan *cryptonightWork, shareChan chan cryptonightShare) {
	hashes := uint32(0)
	startTime := time.Now()

	work := <-workChan
	cl.SetJob(work.input, work.target)

	for {
		select {
		default:
			cl.Nonce = work.nextNonce(cl.Intensity * 16)

			results := make([]uint32, 0x100)

			err := cl.RunJob(results)
			if err != nil {
				log.Errorw("cl error", "error", err)
				return
			}

			for i := uint32(0); i < results[0xFF]; i++ {
				input := work.input
				binary.LittleEndian.PutUint32(input[NonceIndex:], results[i])

				var result []byte
				if w.lite {
					result = hash.CryptonightLite(input)
				} else {
					result = hash.Cryptonight(input)
				}

				if binary.LittleEndian.Uint64(result[24:]) < work.target {
					shareChan <- cryptonightShare{result, results[i]}
				} else {
					log.Errorw("invalid result from CL worker")
				}
			}

			hashes += cl.Intensity
		case newWork, ok := <-workChan:
			w.statMu.Lock()
			w.gpuStats[platform][index] = float32(hashes) / float32(time.Since(startTime).Seconds())
			startTime = time.Now()
			hashes = 0
			w.statMu.Unlock()
			if !ok {
				return
			}
			work = newWork
			cl.SetJob(work.input, work.target)
		}

	}
}

func (w *cryptonight) cpuThread(cpu, index int, workChan chan *cryptonightWork, shareChan chan cryptonightShare) {
	hashes := float32(0)
	startTime := time.Now()

	work := <-workChan

	for {
		select {
		default:
			n := work.nextNonce(64)
			input := work.input

			for i := n; i < n+64; i++ {
				binary.BigEndian.PutUint32(input[NonceIndex:], i)

				var result []byte
				if w.lite {
					result = hash.CryptonightLite(input)
				} else {
					result = hash.Cryptonight(input)
				}

				if binary.LittleEndian.Uint64(result[24:]) < work.target {
					shareChan <- cryptonightShare{result, i}
				}
			}

			hashes += 64
		case newWork, ok := <-workChan:
			w.statMu.Lock()
			w.cpuStats[cpu][index] = float32(hashes) / float32(time.Since(startTime).Seconds())
			startTime = time.Now()
			hashes = 0
			w.statMu.Unlock()
			if !ok {
				return
			}
			work = newWork
		}
	}
}

func (w *cryptonight) Stats() Stats {
	stats := Stats{
		CPUStats: []CPUStats{},
		GPUStats: []GPUStats{},
	}

	w.statMu.RLock()
	defer w.statMu.RUnlock()

	for platform, indexes := range w.gpuStats {
		for index, stat := range indexes {
			stats.GPUStats = append(stats.GPUStats, GPUStats{
				Platform: platform,
				Hashrate: stat,
				Index:    index,
			})
		}
	}

	for cpu, stat := range w.cpuStats {
		hashrate := float32(0)
		for _, hps := range stat {
			hashrate += hps
		}
		stats.CPUStats = append(stats.CPUStats, CPUStats{
			Hashrate: hashrate,
			Index:    cpu,
		})
	}

	return stats
}

func (w *cryptonight) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: true,
		CUDA:   false,
	}
}
