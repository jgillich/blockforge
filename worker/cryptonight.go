package worker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sync/atomic"
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
	clDevices  []CLDeviceConfig
	processors []ProcessorConfig
	stratum    stratum.Client
	lite       bool
	nonce      uint32
	cpuStats   map[int]map[int]float32
	gpuStats   map[int]map[int]float32
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
	closer := make(chan int, 0)

	clWorkers := make([]*CryptonightCLWorker, len(w.clDevices))
	if len(w.clDevices) > 0 {
		for i, device := range w.clDevices {
			worker, err := NewCryptonightCLWorker(device, w.lite)
			if err != nil {
				return err
			}
			clWorkers[i] = worker
		}
	}

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

		atomic.StoreUint32(&w.nonce, 0)

		for i, cl := range clWorkers {
			platformIndex := w.clDevices[i].Device.Platform.Index
			deviceIndex := w.clDevices[i].Device.Index
			if w.gpuStats[platformIndex] == nil {
				w.gpuStats[deviceIndex] = map[int]float32{
					deviceIndex: 0,
				}
			} else {
				w.gpuStats[platformIndex][deviceIndex] = 0
			}

			go func(cl *CryptonightCLWorker) {
				log.Debugf("started cl thread %v/%v", platformIndex, deviceIndex)
				w.gpuStats[platformIndex][deviceIndex] = w.gpuThread(cl, job, target, closer)
				log.Debugf("stopped cl thread %v/%v", platformIndex, deviceIndex)
			}(cl)
		}

		if w.processors != nil {
			for _, conf := range w.processors {
				w.cpuStats[conf.Processor.Index] = map[int]float32{}
				for i := 0; i < conf.Threads; i++ {
					cpuIndex := conf.Processor.Index
					threadIndex := i
					go func() {
						log.Debugf("started cpu thread %v/%v", cpuIndex, threadIndex)
						w.cpuStats[cpuIndex][threadIndex] = w.cpuThread(job, target, closer)
						log.Debugf("stopped cpu thread %v/%v", cpuIndex, threadIndex)
					}()
				}
			}
		}

	}
}

func (w *cryptonight) gpuThread(cl *CryptonightCLWorker, job stratum.Job, target uint64, closer chan int) float32 {
	hashes := uint32(0)
	startTime := time.Now()

	input, err := hex.DecodeString(job.Blob)
	if err != nil {
		log.Errorf("malformed blob: '%v'", job.Blob)
		return 0
	}

	cl.SetJob(input, target)

	for {
		select {
		default:
			cl.Nonce = w.nextNonce(cl.Intensity * 16)
		case <-closer:
			return float32(hashes) / float32(time.Since(startTime).Seconds())
		}
		for i := 0; i < 16; i++ {
			select {
			default:
				results := make([]uint32, 0x100)

				err := cl.RunJob(results)
				if err != nil {
					log.Errorw("cl error", "error", err)
					return 0
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
				return float32(hashes) / float32(time.Since(startTime).Seconds())
			}
		}
	}
}

func (w *cryptonight) cpuThread(job stratum.Job, target uint64, closer chan int) float32 {
	hashes := float32(0)
	startTime := time.Now()

	input, err := hex.DecodeString(job.Blob)
	if err != nil {
		log.Errorf("malformed blob: '%v'", job.Blob)
		return 0
	}

	for {
		select {
		default:
			n := w.nextNonce(16)

			for i := n; i < n+16; i++ {
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
			}

			hashes += 16
		case <-closer:
			return hashes / float32(time.Since(startTime).Seconds())
		}
	}
}

func (w *cryptonight) Stats() Stats {
	stats := Stats{
		CPUStats: []CPUStats{},
		GPUStats: []GPUStats{},
	}

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

func (w *cryptonight) nextNonce(size uint32) uint32 {
	for {
		val := atomic.LoadUint32(&w.nonce)
		if atomic.CompareAndSwapUint32(&w.nonce, val, val+size) {
			return val
		}
	}
}

func (w *cryptonight) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: true,
		CUDA:   false,
	}
}
