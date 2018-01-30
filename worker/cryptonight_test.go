package worker

import (
	"testing"

	"gitlab.com/blockforge/blockforge/hardware/processor"
	"gitlab.com/blockforge/blockforge/log"

	"gitlab.com/blockforge/blockforge/hardware/opencl"

	"gitlab.com/blockforge/blockforge/stratum"
)

func init() {
	log.InitializeTesting()
}

func TestCryptonightCPU(t *testing.T) {
	stratumClient := NewStratumTestClient()
	defer stratumClient.Close()

	processors, err := processor.GetProcessors()
	if err != nil {
		t.Fatal(err)
	}

	worker := NewCryptonight(Config{
		Stratum: stratumClient,
		Processors: []ProcessorConfig{
			ProcessorConfig{Threads: 1, Processor: processors[0]},
		},
	}, false)

	go worker.Work()
	stratumClient.Jobs() <- stratum.Job{
		Blob:   "0606fcfc85d305d5f238078d3eaf897b43bc2548024c4c753c15584cd30a9323be296e1554ecb50000001bd9286f6d087f92749e8f094090af879f8bc16ddce6db71e5ed745e7ed806a98e09",
		Target: "e2361a00",
	}

	share := <-stratumClient.Shares

	if share.Result != "03deef54ac208e5c4c41b608fa3c37436c5350858766d332fffbd8b06efc0700" || share.Nonce != "0000001b" {
		t.Errorf("Invalid share result '%v' / nonce '%v'", share.Result, share.Nonce)
	}
}

func TestCryptonightCL(t *testing.T) {

	// TODO
	if true {
		return
	}

	stratumClient := NewStratumTestClient()
	defer stratumClient.Close()

	platforms, err := opencl.GetPlatforms()
	if err != nil {
		t.Fatal("no opencl platforms found")
	}
	devices, err := opencl.GetDevices(platforms[0])
	if err != nil || len(devices) == 0 {
		t.Fatal("no opencl devices found")
	}

	worker := NewCryptonight(Config{
		Stratum: stratumClient,
		CLDevices: []CLDeviceConfig{
			CLDeviceConfig{
				Intensity: 32,
				Device:    devices[0],
			},
		},
	}, false)

	go worker.Work()

	stratumClient.Jobs() <- stratum.Job{
		Blob:   "0606fcfc85d305d5f238078d3eaf897b43bc2548024c4c753c15584cd30a9323be296e1554ecb500000000d9286f6d087f92749e8f094090af879f8bc16ddce6db71e5ed745e7ed806a98e09",
		Target: "e2361a00",
	}

	share := <-stratumClient.Shares

	if share.Result != "03deef54ac208e5c4c41b608fa3c37436c5350858766d332fffbd8b06efc0700" || share.Nonce != "0000001b" {
		t.Errorf("Invalid share result '%v' / nonce '%v'", share.Result, share.Nonce)
	}
}

func TestCryptonightLite(t *testing.T) {
	stratumClient := NewStratumTestClient()
	defer stratumClient.Close()

	processors, err := processor.GetProcessors()
	if err != nil {
		t.Error(err)
	}

	worker := NewCryptonight(Config{
		Stratum: stratumClient,
		Processors: []ProcessorConfig{
			ProcessorConfig{Threads: 1, Processor: processors[0]},
		},
	}, true)

	go worker.Work()

	stratumClient.Jobs() <- stratum.Job{
		Blob:   "0100c1ee92d3057df2467aa8ad2a8be661a38259a86c2aaf3018b5baffe99e23a335a90633e948000000003557aac2d2b4e74cddaee41c734a65a5d257175761a0cce0a453e4b386cbd87802",
		Target: "26310800",
	}

	share := <-stratumClient.Shares

	if share.Result != "8ff9e89e9f5d6d52fc8fedc7e6e02b8c981cdae34ae4ed09df419c11f7fd0200" || share.Nonce != "0000044e" {
		t.Errorf("Invalid share")
	}
}
