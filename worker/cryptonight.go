package worker

import (
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/jgillich/go-opencl/cl"

	"gitlab.com/jgillich/autominer/hash"
	"gitlab.com/jgillich/autominer/log"

	"gitlab.com/jgillich/autominer/stratum"
)

var CryptonightMemory = 2097152
var CryptonightMask = 0x1FFFF0
var CryptonightIter = 0x80000

var CryptonightLiteMemory = 1048576
var CryptonightLiteMask = 0xFFFF0
var CryptonightLiteIter = 0x40000

// NonceIndex is the starting location of nonce in blob
var NonceIndex = 78

// NonceWidth is the char width of nonce in blob
var NonceWidth = 8

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
	clworker   *CryptonightCLWorker
	cpuThreads int
	processors []ProcessorConfig
	stratum    stratum.Client
	light      bool
	cpuStats   map[int]map[int]float32
}

func NewCryptonight(config Config, light bool) Worker {
	var clworker *CryptonightCLWorker
	if len(config.CLDevices) > 0 {
		clDevices := []*cl.Device{}
		for _, conf := range config.CLDevices {
			clDevices = append(clDevices, conf.Device.CL())
		}

		var err error
		clworker, err = NewCryptonightCLWorker(clDevices)
		if err != nil {
			// TODO
			log.Fatal(err)
		}
	}

	cpuThreads := 0
	if len(config.Processors) > 0 {

		for _, cpu := range config.Processors {
			cpuThreads += cpu.Threads
		}
	}

	return &cryptonight{
		cpuThreads: cpuThreads,
		clworker:   clworker,
		processors: config.Processors,
		stratum:    config.Stratum,
		light:      light,
		cpuStats:   map[int]map[int]float32{},
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

		nounceStepping := uint32(math.MaxUint32 / w.cpuThreads)
		nonce := uint32(0)

		for _, conf := range w.processors {
			w.cpuStats[conf.Processor.Index] = map[int]float32{}
			for i := 0; i < conf.Threads; i++ {
				log.Debugf("starting thread for job '%v'", job.JobId)
				log.Debugf("nonce start '%v' end '%v'", nonce, nonce+nounceStepping)
				go w.cpuThread(conf.Processor.Index, i, job, nonce, nonce+nounceStepping, closer)
				nonce += nounceStepping
			}
		}

	}
}

func (w *cryptonight) cpuThread(cpu int, threadNum int, job stratum.Job, nonceStart uint32, nonceEnd uint32, closer chan int) {
	target := math.MaxUint64 / uint64(math.MaxUint32/hexUint64LE([]byte(job.Target)))
	hashes := float32(0)
	startTime := time.Now()

	defer func() {
		w.cpuStats[cpu][threadNum] = hashes / float32(time.Since(startTime).Seconds())
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

			var result []byte
			if w.light {
				result = hash.CryptonightLite(input)
			} else {
				result = hash.Cryptonight(input)
			}

			val := hexUint64LE([]byte(hex.EncodeToString(result)[48:]))

			if val < target {
				share := stratum.Share{
					MinerId: job.MinerId,
					JobId:   job.JobId,
					Result:  fmt.Sprintf("%x", result),
					Nonce:   fmtNonce(nonce),
				}

				go w.stratum.SubmitShare(&share)
			}
			hashes++
		case <-closer:
			log.Debugf("stopping thread for job '%v'", job.JobId)
			return
		}
	}
}

func (w *cryptonight) Stats() Stats {
	stats := Stats{
		CPUStats: []CPUStats{},
		GPUStats: []GPUStats{},
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
