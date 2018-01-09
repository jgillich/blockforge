package miner

import (
	"fmt"
	"strconv"

	"gitlab.com/jgillich/autominer/coin"
	"gitlab.com/jgillich/autominer/hardware"
)

type Miner struct {
	miners map[string]coin.Miner
}

func New(config Config) (*Miner, error) {
	miner := Miner{
		miners: map[string]coin.Miner{},
	}

	hw, err := hardware.New()
	if err != nil {
		return nil, err
	}

	for coinName, coinConfig := range config.Coins {

		var enabledCPUs []coin.CPUConfig
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

			enabledCPUs = append(enabledCPUs, coin.CPUConfig{
				CPU:     cpu,
				Threads: cpuConfig.Threads,
			})
		}

		var enabledGPUs []coin.GPUConfig
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

			enabledGPUs = append(enabledGPUs, coin.GPUConfig{
				GPU:       gpu,
				Intensity: gpuConfig.Intensity,
			})
		}

		// skip coins with no threads and gpus
		if len(enabledGPUs) == 0 && len(enabledCPUs) == 0 {
			continue
		}

		minerConfig := coin.MinerConfig{
			Coin:   coinName,
			Donate: config.Donate,
			Pool:   coinConfig.Pool,
			CPUSet: enabledCPUs,
			GPUSet: enabledGPUs,
		}

		coin, ok := coin.Coins[coinName]
		if !ok {
			return nil, fmt.Errorf("unsupported coin '%v'", coinName)
		}

		info := coin.Info()

		if !info.SupportsCPU && len(enabledCPUs) > 0 {
			return nil, fmt.Errorf("coin '%v' does not support cpus", coinName)
		}

		for _, gpu := range enabledGPUs {
			if gpu.GPU.Backend == hardware.CUDABackend && !info.SupportsCUDA {
				return nil, fmt.Errorf("coin '%v' does not support CUDA GPU '%v'", coinName, gpu.GPU.Index)
			}
			if gpu.GPU.Backend == hardware.OpenCLBackend && !info.SupportsOpenCL {
				return nil, fmt.Errorf("coin '%v' does not support OpenCL GPU '%v'", coinName, gpu.GPU.Index)
			}
		}

		m, err := coin.Miner(minerConfig)
		if err != nil {
			return nil, err
		}

		miner.miners[coinName] = m
	}

	return &miner, nil
}

func (m *Miner) Start() error {
	for _, miner := range m.miners {
		err := miner.Start()
		if err != nil {
			// shut down previously started miners and return error
			for _, m := range m.miners {
				if m == miner {
					return err
				}
				m.Stop()
			}
		}
	}

	return nil
}

func (m *Miner) Stop() {
	for _, miner := range m.miners {
		miner.Stop()
	}
}

func (m *Miner) Stats() []coin.MinerStats {
	stats := make([]coin.MinerStats, 0)

	for _, miner := range m.miners {
		stats = append(stats, miner.Stats())
	}

	return stats
}
