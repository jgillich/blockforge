package worker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"runtime"
	"time"

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
	clWorkers  []*CryptonightCLWorker
	processors []ProcessorConfig
	stratum    stratum.Client
	lite       bool
	cpuStats   map[int]map[int]float32
}

func NewCryptonight(config Config, lite bool) Worker {
	clWorkers := make([]*CryptonightCLWorker, len(config.CLDevices))
	if len(config.CLDevices) > 0 {
		for i, device := range config.CLDevices {
			worker, err := NewCryptonightCLWorker(device, lite)
			if err != nil {
				// TODO
				log.Fatal(err)
			}
			clWorkers[i] = worker
		}
	}

	return &cryptonight{
		clWorkers:  clWorkers,
		processors: config.Processors,
		stratum:    config.Stratum,
		lite:       lite,
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

		t, err := hex.DecodeString(job.Target)
		if err != nil {
			return err
		}

		var target uint64
		switch len(job.Target) {
		case 8:
			t32 := uint64(binary.LittleEndian.Uint32(t))
			target = math.MaxUint64 / (math.MaxUint32 / t32)
		case 16:
			target = binary.LittleEndian.Uint64(t)
		default:
			return fmt.Errorf("unsupported target length '%v'", len(job.Target))
		}

		log.Infof("job difficulty %v", math.MaxUint64/target)

		if len(w.clWorkers) > 0 {
			go w.gpuThread(w.clWorkers[0], job, target, closer)
		}

		if len(w.processors) > 0 {
			cpuThreads := 0

			for _, cpu := range w.processors {
				cpuThreads += cpu.Threads
			}

			nounceStepping := uint32(math.MaxUint32 / cpuThreads)
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
}

func (w *cryptonight) gpuThread(cl *CryptonightCLWorker, job stratum.Job, target uint64, closer chan int) {
	hashes := uint32(0)
	startTime := time.Now()

	defer func() {
		log.Infof("gpu hashes %v H/s", float64(hashes)/time.Since(startTime).Seconds())
	}()

	input, err := hex.DecodeString(job.Blob)
	if err != nil {
		log.Errorf("malformed blob: '%v'", job.Blob)
		return
	}

	cl.SetJob(input, target)

	for {
		select {
		default:
			results := make([]uint32, 0x100)

			err := cl.RunJob(results)
			if err != nil {
				log.Errorw("cl error", "error", err)
				return
			}

			for i := uint32(0); i < results[0xFF]; i++ {
				nonce := make([]byte, NonceWidth)
				binary.LittleEndian.PutUint32(nonce, results[i])
				for i := 0; i < len(nonce); i++ {
					input[NonceIndex+i] = nonce[i]
				}

				var result []byte
				if w.lite {
					result = hash.CryptonightLite(input)
				} else {
					result = hash.Cryptonight(input)
				}

				if binary.LittleEndian.Uint64(result[24:]) < target {
					share := stratum.Share{
						MinerId: job.MinerId,
						JobId:   job.JobId,
						Result:  fmt.Sprintf("%x", result),
						Nonce:   fmt.Sprintf("%08x", binary.BigEndian.Uint32(nonce)),
					}

					go w.stratum.SubmitShare(&share)
				} else {
					log.Errorw("invalid result from CL worker",
						"blob", job.Blob,
						"target", job.Target,
						"nonce", fmt.Sprintf("%x", nonce))
				}
			}

			hashes += cl.Intensity
		case <-closer:
			log.Debugf("stopping gpu for job '%v'", job.JobId)
			return
		}
	}
}

func (w *cryptonight) cpuThread(cpu int, threadNum int, job stratum.Job, nonceStart uint32, nonceEnd uint32, target uint64, closer chan int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hashes := float32(0)
	startTime := time.Now()

	input, err := hex.DecodeString(job.Blob)
	if err != nil {
		log.Errorf("malformed blob: '%v'", job.Blob)
		return
	}

	defer func() {
		w.cpuStats[cpu][threadNum] = hashes / float32(time.Since(startTime).Seconds())
	}()

	for i := nonceStart; i < nonceEnd; i++ {
		select {
		default:
			nonce := make([]byte, NonceWidth)
			binary.BigEndian.PutUint32(nonce, i)
			for i := 0; i < len(nonce); i++ {
				input[NonceIndex+i] = nonce[i]
			}

			var result []byte
			if w.lite {
				result = hash.CryptonightLite(input)
			} else {
				result = hash.Cryptonight(input)
			}

			if binary.LittleEndian.Uint64(result[24:]) < target {
				share := stratum.Share{
					MinerId: job.MinerId,
					JobId:   job.JobId,
					Result:  fmt.Sprintf("%x", result),
					Nonce:   fmt.Sprintf("%08x", nonce),
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
