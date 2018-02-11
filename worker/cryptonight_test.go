package worker

import (
	"testing"

	"gitlab.com/blockforge/blockforge/algo/cryptonight"
	"gitlab.com/blockforge/blockforge/hardware/processor"
	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	log.InitializeTesting()
}

func cryptonightTestWorker(t *testing.T) (chan *cryptonight.Work, chan cryptonight.Share) {
	work := make(chan *cryptonight.Work, 1)
	shares := make(chan cryptonight.Share, 1)

	processors, err := processor.GetProcessors()
	if err != nil {
		t.Fatal(err)
	}

	worker := Cryptonight{
		Work:   work,
		Shares: shares,
		processors: []ProcessorConfig{
			ProcessorConfig{Threads: 1, Processor: processors[0]},
		},
	}

	go worker.Start()

	return work, shares
}

func TestCryptonight(t *testing.T) {
	// TODO
}

func TestCryptonightLite(t *testing.T) {
	// TODO

}
