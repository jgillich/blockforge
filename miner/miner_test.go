package miner

import (
	"testing"
	"time"

	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	log.InitializeTesting()
}

func TestMiner(t *testing.T) {
	config := Config{
		Coins: defaultCoinConfig,
		Processors: []Processor{
			Processor{
				Coin:    "XMR",
				Enable:  true,
				Index:   0,
				Threads: 1,
			},
		},
	}

	miner, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Minute * 2)

		if miner.Stats()["worker.cpu.0.0"] < 5 {
			t.Logf("extremely low stats")
			t.Fail()
		}

		miner.Stop()
	}()

	err = miner.Start()
	if err != nil {
		t.Fatal(err)
	}

}
