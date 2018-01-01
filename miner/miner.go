package miner

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	config Config
}

func New(config Config) *Miner {
	miner := Miner{config: config}

	return &miner
}

func (m *Miner) Start() error {

	for coinName, coinConfig := range m.config.Coins {

		threads := 0
		for _, cpuConfig := range m.config.CPUs {
			if cpuConfig.Coin == coinName {
				if cpuConfig.Threads <= 0 {
					return errors.New("CPU threads must be 1 or larger")
				}
				threads += cpuConfig.Threads
			}
		}
		if threads > runtime.NumCPU() {
			return errors.New("CPU threads cannot exceed total CPU cores")
		}

		var gpus []int
		for _, gpuConfig := range m.config.GPUs {
			if gpuConfig.Coin == coinName {
				gpus = append(gpus, gpuConfig.Index)
			}
		}

		// skip coins with no threads and gpus
		if len(gpus) == 0 && threads == 0 {
			continue
		}

		mineConfig := coin.MineConfig{
			Donate:     m.config.Donate,
			PoolURL:    coinConfig.Pool.URL,
			PoolUser:   coinConfig.Pool.User,
			PoolPass:   coinConfig.Pool.Pass,
			Threads:    threads,
			GPUIndexes: gpus,
		}

		coin, ok := coin.Coins[coinName]
		if !ok {
			return fmt.Errorf("unsupported coin '%v'", coinName)
		}

		err := coin.Mine(mineConfig)
		if err != nil {
			return err
		}
	}

	// wait for interrupt signal
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	defer signal.Stop(sigc)
	<-sigc

	return nil

}
