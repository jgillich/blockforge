package worker

import (
	"fmt"
	"time"

	metrics "github.com/armon/go-metrics"
	"gitlab.com/blockforge/blockforge/algo/cryptonight"
	"gitlab.com/blockforge/blockforge/log"
)

type Cryptonight struct {
	Algo   *cryptonight.Algo
	Work   <-chan *cryptonight.Work
	Shares chan<- cryptonight.Share

	clDevices  []CLDeviceConfig
	processors []ProcessorConfig
	metrics    *metrics.Metrics
}

func (worker *Cryptonight) Configure(config Config) error {
	worker.clDevices = config.CLDevices
	worker.processors = config.Processors
	worker.metrics = config.Metrics
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

	if len(worker.clDevices) > 0 {
		for i, d := range worker.clDevices {
			cl, err := newCryptonightCL(d, worker.Algo.Lite)
			if err != nil {
				return err
			}
			key := []string{"opencl", fmt.Sprintf("%v", d.Device.Platform.Index), fmt.Sprintf("%v", d.Device.Index)}

			go worker.gpuThread(key, cl, workChannels[i])
		}
	}

	for cpuIndex, conf := range worker.processors {
		for i := 0; i < conf.Threads; i++ {
			key := []string{"cpu", fmt.Sprintf("%v", cpuIndex), fmt.Sprintf("%v", i)}
			go worker.cpuThread(key, workChannels[len(worker.clDevices)+i])
		}
	}

	for work := range worker.Work {
		for _, ch := range workChannels {
			ch <- work
		}
	}

	return nil
}

func (worker *Cryptonight) gpuThread(key []string, cl *cryptonightCL, workChan chan *cryptonight.Work) {
	defer cl.Release()

	var ok bool
	work := <-workChan
	cl.SetJob(work.Input, work.Target)

	for {
		select {
		default:
			start := time.Now()
			results := make([]uint32, 0x100)

			if err := cl.RunJob(results, work.NextNonce(cl.Intensity)); err != nil {
				log.Errorw("cl error", "error", err)
				return
			}

			// number of results is stored in last item of results array
			for i := uint32(0); i < results[0xFF]; i++ {
				if !work.VerifySend(worker.Algo.Lite, results[i], worker.Shares) {
					log.Errorw("invalid result from CL worker")
				}
			}

			worker.metrics.IncrCounter(key, float32(float64(cl.Intensity)/time.Since(start).Seconds()))
		case work, ok = <-workChan:
			if !ok {
				return
			}
			cl.SetJob(work.Input, work.Target)
		}

	}
}

func (worker *Cryptonight) cpuThread(key []string, workChan chan *cryptonight.Work) {
	work := <-workChan
	var ok bool

	for {
		select {
		default:
			start := time.Now()
			work.VerifyRange(worker.Algo.Lite, 64, worker.Shares)
			worker.metrics.IncrCounter(key, float32(64/time.Since(start).Seconds()))
		case work, ok = <-workChan:
			if !ok {
				return
			}
		}
	}
}

func (worker *Cryptonight) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: true,
		CUDA:   false,
	}
}
