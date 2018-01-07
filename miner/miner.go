package miner

import (
	"errors"
	"fmt"
	"runtime"

	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	miners map[string]coin.Miner
}

func New(config Config) (*Miner, error) {
	miner := Miner{
		miners: map[string]coin.Miner{},
	}

	for coinName, coinConfig := range config.Coins {

		threads := 0
		for _, cpuConfig := range config.CPUs {
			if cpuConfig.Coin == coinName {
				if cpuConfig.Threads <= 0 {
					return nil, errors.New("CPU threads must be 1 or larger")
				}
				threads += cpuConfig.Threads
			}
		}
		if threads > runtime.NumCPU() {
			return nil, errors.New("CPU threads cannot exceed total CPU cores")
		}

		var gpus []int
		for _, gpuConfig := range config.GPUs {
			if gpuConfig.Coin == coinName {
				gpus = append(gpus, gpuConfig.Index)
			}
		}

		// skip coins with no threads and gpus
		if len(gpus) == 0 && threads == 0 {
			continue
		}

		minerConfig := coin.MinerConfig{
			Coin:       coinName,
			Donate:     config.Donate,
			PoolURL:    coinConfig.Pool.URL,
			PoolUser:   coinConfig.Pool.User,
			PoolPass:   coinConfig.Pool.Pass,
			PoolEmail:  coinConfig.Pool.Email,
			Threads:    threads,
			GPUIndexes: gpus,
		}

		coin, ok := coin.Coins[coinName]
		if !ok {
			return nil, fmt.Errorf("unsupported coin '%v'", coinName)
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
