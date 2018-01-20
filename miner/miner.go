package miner

import (
	"fmt"
	"log"
	"strconv"

	"gitlab.com/jgillich/autominer/stratum"

	"gitlab.com/jgillich/autominer/hardware"
	"gitlab.com/jgillich/autominer/worker"
)

type Miner struct {
	stratums map[string]stratum.Client
	workers  map[string]worker.Worker
}

func New(config Config) (*Miner, error) {
	miner := Miner{
		stratums: map[string]stratum.Client{},
		workers:  map[string]worker.Worker{},
	}

	hw, err := hardware.New()
	if err != nil {
		return nil, err
	}

	for coinName, coinConfig := range config.Coins {

		var enabledCPUs []worker.CPUConfig
		for index, cpuConfig := range config.CPUs {
			if cpuConfig.Coin != coinName {
				continue
			}
			if cpuConfig.Threads == 0 {
				continue
			}

			cpuIndex, err := strconv.Atoi(index)
			if err != nil {
				return nil, fmt.Errorf("non numeric cpu index '%v'", index)
			}

			var cpu hardware.CPU
			for _, c := range hw.CPUs {
				if c.Index == cpuIndex {
					cpu = c
					break
				}
			}
			if (cpu == hardware.CPU{}) {
				return nil, fmt.Errorf("cpu with index '%v' not found", cpuIndex)
			}

			if cpu.VirtualCores < cpuConfig.Threads {
				return nil, fmt.Errorf("thread count '%v' for cpu '%v' cannot be larger than number of virtual cores '%v'", cpuConfig.Threads, index, cpu.VirtualCores)
			}

			enabledCPUs = append(enabledCPUs, worker.CPUConfig{
				CPU:     cpu,
				Threads: cpuConfig.Threads,
			})
		}

		var enabledGPUs []worker.GPUConfig
		for index, gpuConfig := range config.GPUs {
			if gpuConfig.Coin != coinName {
				continue
			}

			gpuIndex, err := strconv.Atoi(index)
			if err != nil {
				return nil, fmt.Errorf("non numeric gpu index '%v'", index)
			}

			var gpu hardware.GPU
			for _, c := range hw.GPUs {
				if c.Index == gpuIndex {
					gpu = c
					break
				}
			}
			if (gpu == hardware.GPU{}) {
				return nil, fmt.Errorf("gpu with index '%v' not found", gpuIndex)
			}

			enabledGPUs = append(enabledGPUs, worker.GPUConfig{
				GPU:       gpu,
				Intensity: gpuConfig.Intensity,
			})
		}

		// skip coins with no threads and gpus
		if len(enabledGPUs) == 0 && len(enabledCPUs) == 0 {
			continue
		}

		stratum, err := stratum.NewClient("jsonrpc", coinConfig.Pool)
		if err != nil {
			return nil, err
		}

		workerConfig := worker.Config{
			Stratum: stratum,
			Donate:  config.Donate,
			CPUSet:  enabledCPUs,
			GPUSet:  enabledGPUs,
		}

		worker, err := worker.New(coinName, workerConfig)
		if err != nil {
			return nil, err
		}

		capabilities := worker.Capabilities()

		if !capabilities.CPU && len(enabledCPUs) > 0 {
			return nil, fmt.Errorf("coin '%v' does not support cpus", coinName)
		}

		for _, gpu := range enabledGPUs {
			if gpu.GPU.Backend == hardware.CUDABackend && !capabilities.CUDA {
				return nil, fmt.Errorf("coin '%v' does not support CUDA GPU '%v'", coinName, gpu.GPU.Index)
			}
			if gpu.GPU.Backend == hardware.OpenCLBackend && !capabilities.OpenCL {
				return nil, fmt.Errorf("coin '%v' does not support OpenCL GPU '%v'", coinName, gpu.GPU.Index)
			}
		}

		go func() {
			err := worker.Work()
			if err != nil {
				log.Fatal(err)
			}
		}()

		miner.stratums[coinName] = stratum
		miner.workers[coinName] = worker
	}

	return &miner, nil
}

func (m *Miner) Stop() {
	for _, stratum := range m.stratums {
		stratum.Close()
	}
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
