package miner

import (
	"fmt"
	"time"

	metrics "github.com/armon/go-metrics"
	"gitlab.com/blockforge/blockforge/coin"
	"gitlab.com/blockforge/blockforge/hardware/opencl"
	"gitlab.com/blockforge/blockforge/hardware/processor"
	"gitlab.com/blockforge/blockforge/log"
	"gitlab.com/blockforge/blockforge/stratum"
	"gitlab.com/blockforge/blockforge/worker"
)

type Miner struct {
	stratums map[string]stratum.Client
	workers  map[string]worker.Worker
	err      chan error
	sink     *metrics.InmemSink
	metrics  *metrics.Metrics
}

func New(config *Config) (*Miner, error) {
	sink := metrics.NewInmemSink(time.Minute, 10*time.Minute)

	metrics, err := metrics.New(metrics.DefaultConfig("worker"), sink)
	if err != nil {
		return nil, err
	}

	miner := Miner{
		stratums: map[string]stratum.Client{},
		workers:  map[string]worker.Worker{},
		err:      make(chan error),
		sink:     sink,
		metrics:  metrics,
	}

	processors, err := processor.GetProcessors()
	if err != nil {
		return nil, err
	}

	var clPlatforms []*opencl.Platform
	if len(config.OpenCLDevices) > 0 {
		clPlatforms, err = opencl.GetPlatforms()
		if err != nil {
			return nil, err
		}
	}

	for name, coinConfig := range config.Coins {

		coin := coin.Lookup(name)
		if coin == nil {
			return nil, fmt.Errorf("coin '%v' is not supported", name)
		}

		var pConf []worker.ProcessorConfig
		for _, conf := range config.Processors {
			if !conf.Enable || conf.Coin != name {
				continue
			}

			var processor *processor.Processor
			for _, p := range processors {
				if p.Index == conf.Index {
					processor = p
					break
				}
			}

			if processor == nil {
				return nil, fmt.Errorf("cpu index '%v' does not exist", conf.Index)
			}

			if conf.Threads > processor.VirtualCores {
				return nil, fmt.Errorf("threads for cpu '%v' cannot be higher than virtual cores (%v > %v)", conf.Index, conf.Threads, processor.VirtualCores)
			}

			pConf = append(pConf, worker.ProcessorConfig{Threads: conf.Threads, Processor: processor})
		}

		var clConf []worker.CLDeviceConfig
		for _, conf := range config.OpenCLDevices {
			if !conf.Enable || conf.Coin != name {
				continue
			}

			var device *opencl.Device
			for _, p := range clPlatforms {
				if p.Index == conf.Platform {
					for _, d := range p.Devices {
						if d.Index == conf.Index {
							device = d
							break
						}
					}
					break
				}
			}

			if device == nil {
				return nil, fmt.Errorf("opencl device platform '%v' index '%v' does not exist", conf.Platform, conf.Index)
			}

			clConf = append(clConf, worker.CLDeviceConfig{
				Intensity: conf.Intensity,
				Worksize:  conf.Worksize,
				Device:    device,
			})
		}

		// skip coins without workers
		if (len(pConf) + len(clConf)) == 0 {
			continue
		}

		stratum, err := stratum.NewClient(coin.Protocol, coinConfig.Pool)
		if err != nil {
			return nil, err
		}

		workerConfig := worker.Config{
			Donate:     config.Donate,
			Processors: pConf,
			CLDevices:  clConf,
			Metrics:    miner.metrics,
		}

		worker := stratum.Worker(coin.Algo)
		if err := worker.Configure(workerConfig); err != nil {
			return nil, err
		}

		capabilities := worker.Capabilities()

		if !capabilities.CPU && len(pConf) > 0 {
			return nil, fmt.Errorf("coin '%v' does not support processors", name)
		}

		if !capabilities.OpenCL && len(clConf) > 0 {
			return nil, fmt.Errorf("coin '%v' does not support opencl devices", name)
		}

		miner.stratums[name] = stratum
		miner.workers[name] = worker
	}
	return &miner, nil
}

func (miner *Miner) Start() error {
	for _, w := range miner.workers {
		go func(worker worker.Worker) {
			err := worker.Start()
			if err != nil {
				miner.err <- err
			}
		}(w)
	}

	log.Debug("miner started")
	defer log.Debug("miner stopped")
	return <-miner.err
}

func (miner *Miner) Stop() {
	for _, stratum := range miner.stratums {
		stratum.Close()
	}
	miner.err <- nil
	close(miner.err)
}

func (miner *Miner) Stats() map[string]float64 {
	stats := map[string]float64{}
	data := miner.sink.Data()
	metrics := data[len(data)-1]
	metrics.RLock()
	defer metrics.RUnlock()
	for key, counter := range metrics.Counters {
		stats[key] = counter.AggregateSample.Mean()
	}
	return stats
}
