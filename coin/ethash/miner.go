package ethash

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/node"

	"github.com/ethereum/go-ethereum/eth"
	"gitlab.com/jgillich/autominer/cgo/ethminer"
	"gitlab.com/jgillich/autominer/coin"
)

type Miner struct {
	coin string
	cpu  *eth.Ethereum
	gpu  ethminer.Ethminer
}

func NewMiner(config coin.MinerConfig) (coin.Miner, error) {

	miner := Miner{coin: config.Coin}

	if config.Threads > 0 {
		nodeConfig := node.DefaultConfig
		ethConfig := eth.DefaultConfig

		ethConfig.MinerThreads = config.Threads
		ethConfig.Etherbase = common.HexToAddress(config.PoolUser)

		stack, err := node.New(&nodeConfig)
		if err != nil {
			return nil, fmt.Errorf("Failed to create the protocol stack: %v", err)
		}

		// register eth service
		err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
			return eth.New(ctx, &ethConfig)
		})
		if err != nil {
			return nil, fmt.Errorf("Failed to register the Ethereum service: %v", err)
		}

		if err := stack.Start(); err != nil {
			return nil, fmt.Errorf("Error starting protocol stack: %v", err)
		}

		// Mining only makes sense if a full Ethereum node is running
		if err := stack.Service(&miner.cpu); err != nil {
			return nil, fmt.Errorf("ethereum service not running: %v", err)
		}

		type threaded interface {
			SetThreads(threads int)
		}
		miner.cpu.Engine().(threaded).SetThreads(config.Threads)

		go func() {
			sigc := make(chan os.Signal, 1)
			signal.Notify(sigc, os.Interrupt)
			defer signal.Stop(sigc)
			<-sigc
			go stack.Stop()
		}()

	}

	if len(config.GPUIndexes) > 0 {
		miner.gpu = ethminer.NewEthminer()

		u, err := url.Parse(config.PoolURL)
		if err != nil {
			return nil, err
		}

		openclDevices := make([]uint, len(config.GPUIndexes))
		for i, idx := range config.GPUIndexes {
			openclDevices[i] = uint(idx)
		}

		miner.gpu.SetM_farmURL(u.Hostname())
		miner.gpu.SetM_user(config.PoolUser)
		miner.gpu.SetM_pass(config.PoolPass)
		miner.gpu.SetM_port(u.Port())
		miner.gpu.SetM_openclDevices(&openclDevices[0])
		miner.gpu.SetM_openclDeviceCount(uint(len(openclDevices)))

	}

	return &miner, nil
}

func (m *Miner) Start() error {
	if m.cpu != nil {
		if err := m.cpu.StartMining(true); err != nil {
			return fmt.Errorf("Failed to start mining: %v", err)
		}
	}

	if m.gpu != nil {
		go m.gpu.Start()
	}

	return nil
}

func (m *Miner) Stats() coin.MinerStats {
	hashrate := 0
	if m.cpu != nil {
		hashrate += int(m.cpu.Miner().HashRate())
	}
	if m.gpu != nil {
		hashrate += m.gpu.GetM_hashrate()
	}
	return coin.MinerStats{
		Coin:     m.coin,
		Hashrate: float32(hashrate),
	}
}
