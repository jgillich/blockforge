package miner

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	log.InitializeTesting()
}

func TestMiner(t *testing.T) {

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	go func() {
		for {
			var buf bytes.Buffer
			written, err := io.Copy(&buf, r)

			if err != nil {
				t.Fatal(err)
			}

			if written > 0 {
				t.Fail()
			}
		}

	}()

	config := Config{
		Coins: DefaultCoinConfig,
		Processors: []Processor{
			Processor{
				Coin:    "XMR",
				Enable:  true,
				Index:   0,
				Threads: 2,
			},
		},
	}

	miner, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Minute)

		stats := miner.Stats()

		for _, stat := range stats.CPUStats {
			t.Logf("CPU %v: %.2f H/s", stat.Index, stat.Hashrate)
		}

		if stats.CPUStats[0].Hashrate < 10 {
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
