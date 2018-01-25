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
	clWorkers  []*CryptonightCLWorker
	cpuThreads int
	processors []ProcessorConfig
	stratum    stratum.Client
	light      bool
	cpuStats   map[int]map[int]float32
}

func NewCryptonight(config Config, light bool) Worker {
	clWorkers := make([]*CryptonightCLWorker, len(config.CLDevices))
	if len(config.CLDevices) > 0 {
		for i, device := range config.CLDevices {
			worker, err := NewCryptonightCLWorker(device, light)
			if err != nil {
				// TODO
				log.Fatal(err)
			}
			clWorkers[i] = worker
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
		clWorkers:  clWorkers,
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

		target := math.MaxUint64 / uint64(math.MaxUint32/hexUint64LE([]byte(job.Target)))

		if len(w.clWorkers) > 0 {
			w.gpuThread(w.clWorkers[0], job, target)
		}

		nounceStepping := uint32(math.MaxUint32 / w.cpuThreads)
		nonce := uint32(0)

		for _, conf := range w.processors {
			w.cpuStats[conf.Processor.Index] = map[int]float32{}
			for i := 0; i < conf.Threads; i++ {
				log.Debugf("starting thread for job '%v'", job.JobId)
				log.Debugf("nonce start '%v' end '%v'", nonce, nonce+nounceStepping)
				go w.cpuThread(conf.Processor.Index, i, job, nonce, nonce+nounceStepping, target, closer)
				nonce += nounceStepping
			}
		}

	}
}

func (w *cryptonight) gpuThread(cl *CryptonightCLWorker, job stratum.Job, target uint64) {
	hashes := 0

	input, err := hex.DecodeString(job.Blob)
	if err != nil {
		log.Errorf("malformed blob: '%v'", job.Blob)
		return
	}

	cl.SetJob(input, target)

	for {
		results, err := cl.RunJob()
		if err != nil {
			log.Errorw("cl error", "error", err)
			return
		}

		for i := byte(0); i < results[0xFF]; i++ {
			panic("???")
			/*
				uint8_t	bWorkBlob[112];
				uint8_t	bResult[32];

				memcpy(bWorkBlob, oWork.bWorkBlob, oWork.iWorkSize);
				memset(bResult, 0, sizeof(job_result::bResult));

				*(uint32_t*)(bWorkBlob + 39) = results[i];

				hash_fun(bWorkBlob, oWork.iWorkSize, bResult, cpu_ctx);
				if ( (*((uint64_t*)(bResult + 24))) < oWork.iTarget)
					executor::inst()->push_event(ex_event(job_result(oWork.sJobID, results[i], bResult, iThreadNo), oWork.iPoolId));
				else
					executor::inst()->push_event(ex_event("AMD Invalid Result", pGpuCtx->deviceIdx, oWork.iPoolId));*/
		}

		continue

		val := hexUint64LE([]byte(hex.EncodeToString(results)[48:]))

		if val < target {
			share := stratum.Share{
				MinerId: job.MinerId,
				JobId:   job.JobId,
				Result:  fmt.Sprintf("%x", results),
				Nonce:   fmtNonce(uint32(cl.Nonce)),
			}

			go w.stratum.SubmitShare(&share)
		}
		hashes += cl.Intensity
	}
}

func (w *cryptonight) cpuThread(cpu int, threadNum int, job stratum.Job, nonceStart uint32, nonceEnd uint32, target uint64, closer chan int) {
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
