package miner

import (
	"fmt"

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
}

func New(config Config) (*Miner, error) {
	miner := Miner{
		stratums: map[string]stratum.Client{},
		workers:  map[string]worker.Worker{},
		err:      make(chan error),
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

	for name, coin := range config.Coins {

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

			pConf = append(pConf, worker.ProcessorConfig{conf.Threads, processor})
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

			clConf = append(clConf, worker.CLDeviceConfig{conf.Intensity, device})
		}

		// skip coins without workers
		if (len(pConf) + len(clConf)) == 0 {
			continue
		}

		stratum, err := stratum.NewClient("jsonrpc", coin.Pool)
		if err != nil {
			return nil, err
		}

		workerConfig := worker.Config{
			Stratum:    stratum,
			Donate:     config.Donate,
			Processors: pConf,
			CLDevices:  clConf,
		}

		worker, err := worker.New(name, workerConfig)
		if err != nil {
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

func (m *Miner) Start() error {
	for _, worker := range m.workers {
		go func() {
			err := worker.Work()
			if err != nil {
				m.err <- err
			}
		}()
	}
	log.Debug("miner started")
	defer log.Debug("miner stopped")
	return <-m.err
}

func (m *Miner) Stop() {
	for _, stratum := range m.stratums {
		stratum.Close()
	}
	m.err <- nil
	close(m.err)
}

func (m *Miner) Stats() worker.Stats {
	stats := worker.Stats{
		CPUStats: []worker.CPUStats{},
		GPUStats: []worker.GPUStats{},
	}

	for _, worker := range m.workers {
		s := worker.Stats()

		for _, cpuStat := range s.CPUStats {
			stats.CPUStats = append(stats.CPUStats, cpuStat)
		}

		for _, gpuStat := range s.GPUStats {
			stats.GPUStats = append(stats.GPUStats, gpuStat)
		}
	}

	return stats
}
