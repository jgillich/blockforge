package etherum

import (
	"log"
	"os"
	"os/signal"

	"gitlab.com/jgillich/autominer/coin"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
)

func (e *Ethereum) mineCPU(config coin.CoinConfig) {

	nodeConfig := node.DefaultConfig
	ethConfig := eth.DefaultConfig

	stack, err := node.New(&nodeConfig)
	if err != nil {
		log.Fatalf("Failed to create the protocol stack: %v", err)
	}

	//ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	// TODO retrieve etherbase from account
	// https://github.com/ethereum/go-ethereum/blob/46e5583993afe7b9d0ff432f846b2a97bcb89876/cmd/utils/flags.go#L764
	ethConfig.Etherbase = common.HexToAddress("0x7ef5a6135f1fd6a02593eedc869c6d41d934aef8")

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

	// TODO threads from config
	ethereum.Engine().(threaded).SetThreads(1)

	// TODO config option
	// Set the gas price to the limits from the CLI and start mining
	//ethereum.TxPool().SetGasPrice(utils.GlobalBig(ctx, utils.GasPriceFlag.Name))

	if err := ethereum.StartMining(true); err != nil {
		log.Fatalf("Failed to start mining: %v", err)
	}

	stack.Wait()
}
