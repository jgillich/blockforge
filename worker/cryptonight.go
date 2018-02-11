package worker

import (
	"sync"
	"time"

	"gitlab.com/blockforge/blockforge/algo/cryptonight"
	"gitlab.com/blockforge/blockforge/log"
)

type Cryptonight struct {
	Lite   bool
	Work   <-chan *cryptonight.Work
	Shares chan<- cryptonight.Share

	clDevices  []CLDeviceConfig
	processors []ProcessorConfig
	cpuStats   map[int]map[int]float32
	gpuStats   map[int]map[int]float32
	statMu     sync.RWMutex
}

func (worker *Cryptonight) Configure(config Config) error {
	worker.clDevices = config.CLDevices
	worker.processors = config.Processors
	worker.cpuStats = map[int]map[int]float32{}
	worker.gpuStats = map[int]map[int]float32{}
	return nil
}

func (worker *Cryptonight) Start() error {
	totalThreads := len(worker.clDevices)
	for _, c := range worker.processors {
		totalThreads += c.Threads
	}

	workChannels := make([]chan *cryptonight.Work, totalThreads)
	for i := 0; i < totalThreads; i++ {
		workChannels[i] = make(chan *cryptonight.Work, 1)
		index := i
		defer close(workChannels[index])
	}

	worker.statMu.Lock()
	{
		if len(worker.clDevices) > 0 {
			for i, d := range worker.clDevices {
				worker.gpuStats[d.Device.Platform.Index] = map[int]float32{}
				cl, err := newCryptonightCLWorker(d, worker.Lite)
				if err != nil {
					return err
				}
				go worker.gpuThread(d.Device.Platform.Index, d.Device.Index, cl, workChannels[i])
			}
		}

		for cpuIndex, conf := range worker.processors {
			worker.cpuStats[cpuIndex] = map[int]float32{}
			for i := 0; i < conf.Threads; i++ {
				go worker.cpuThread(cpuIndex, i, workChannels[len(worker.clDevices)+i])
			}
		}
	}
	worker.statMu.Unlock()

	for work := range worker.Work {
		for _, ch := range workChannels {
			ch <- work
		}
	}

	return nil
}

func (worker *Cryptonight) gpuThread(platform, index int, cl *cryptonightCLWorker, workChan chan *cryptonight.Work) {
	log.Debugf("gpu thread %v/%v started", platform, index)
	defer log.Debugf("gpu thread %v/%v stopped", platform, index)
	defer cl.Release()

	hashes := uint32(0)
	start := time.Now()

	work := <-workChan
	cl.SetJob(work.Input, work.Target)

	for {
		select {
		default:
			results := make([]uint32, 0x100)

			err := cl.RunJob(results, work.NextNonce(cl.Intensity))
			if err != nil {
				log.Errorw("cl error", "error", err)
				return
			}

			for i := uint32(0); i < results[0xFF]; i++ {
				if !work.VerifySend(worker.Lite, results[i], worker.Shares) {
					log.Errorw("invalid result from CL worker")
				}
			}

			hashes += cl.Intensity
		case newWork, ok := <-workChan:
			elapsed := time.Since(start).Seconds()
			if elapsed > 0 {
				worker.statMu.Lock()
				worker.gpuStats[platform][index] = float32(hashes) / float32(elapsed)
				worker.statMu.Unlock()
				start = time.Now()
				hashes = 0
			}
			if !ok {
				return
			}
			work = newWork
			cl.SetJob(work.Input, work.Target)
		}

	}
}

func (worker *Cryptonight) cpuThread(cpu, index int, workChan chan *cryptonight.Work) {
	log.Debugf("cpu thread %v/%v started", cpu, index)
	defer log.Debugf("cpu thread %v/%v stopped", cpu, index)

	hashes := 0
	start := time.Now()

	work := <-workChan
	var ok bool

	for {
		select {
		default:
			work.VerifyRange(worker.Lite, 64, worker.Shares)
			hashes += 64
		case work, ok = <-workChan:
			elapsed := time.Since(start).Seconds()
			if elapsed > 0 {
				worker.statMu.Lock()
				worker.cpuStats[cpu][index] = float32(hashes) / float32(elapsed)
				worker.statMu.Unlock()
				start = time.Now()
				hashes = 0
			}
			if !ok {
				return
			}
		}
	}
}

func (worker *Cryptonight) Stats() Stats {
	stats := Stats{
		CPUStats: []CPUStats{},
		GPUStats: []GPUStats{},
	}

	worker.statMu.RLock()
	defer worker.statMu.RUnlock()

	for platform, indexes := range worker.gpuStats {
		for index, stat := range indexes {
			stats.GPUStats = append(stats.GPUStats, GPUStats{
				Platform: platform,
				Hashrate: stat,
				Index:    index,
			})
		}
	}

	for cpu, stat := range worker.cpuStats {
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

func (worker *Cryptonight) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: true,
		CUDA:   false,
	}
}
