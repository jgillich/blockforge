package etherum

import (
	"log"
	"os"
	"os/signal"
	"time"

	"gitlab.com/jgillich/autominer/coin"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
)

func (e *Ethash) mineCPU(config coin.MineConfig) {

	nodeConfig := node.DefaultConfig
	ethConfig := eth.DefaultConfig

	ethConfig.MinerThreads = config.Threads
	ethConfig.Etherbase = common.HexToAddress(config.PoolUser)

	stack, err := node.New(&nodeConfig)
	if err != nil {
		log.Fatalf("Failed to create the protocol stack: %v", err)
	}

	// register eth service
	err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return eth.New(ctx, &ethConfig)
	})
	if err != nil {
		log.Fatalf("Failed to register the Ethereum service: %v", err)
	}

	if err := stack.Start(); err != nil {
		log.Fatalf("Error starting protocol stack: %v", err)
	}

	// Mining only makes sense if a full Ethereum node is running
	var ethereum *eth.Ethereum
	if err := stack.Service(&ethereum); err != nil {
		log.Fatalf("ethereum service not running: %v", err)
	}

	type threaded interface {
		SetThreads(threads int)
	}

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt)
		defer signal.Stop(sigc)
		<-sigc
		log.Println("Got interrupt, shutting down...")
		go stack.Stop()
	}()

	ethereum.Engine().(threaded).SetThreads(config.Threads)

	// TODO config option
	// Set the gas price to the limits from the CLI and start mining
	//ethereum.TxPool().SetGasPrice(utils.GlobalBig(ctx, utils.GasPriceFlag.Name))

	if err := ethereum.StartMining(true); err != nil {
		log.Fatalf("Failed to start mining: %v", err)
	}

	go func() {
		for {
			config.Stats <- coin.MineStats{
				Coin:     config.Coin,
				Hashrate: float32(ethereum.Miner().HashRate()),
			}
			time.Sleep(time.Second * 10)
		}
	}()

	stack.Wait()
}
