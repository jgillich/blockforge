package worker

import (
	"testing"

	"gitlab.com/jgillich/autominer/stratum"
)

func TestCryptonote(t *testing.T) {
	stratumClient := NewStratumTestClient()

	worker := NewCryptonight(Config{
		Stratum: stratumClient,
		CPUSet: []CPUConfig{
			CPUConfig{Threads: 1},
		},
	}, false)

	go worker.Work()

	// nonce 0000001b
	stratumClient.Jobs() <- stratum.Job{
		Blob:   "0606fcfc85d305d5f238078d3eaf897b43bc2548024c4c753c15584cd30a9323be296e1554ecb50000001bd9286f6d087f92749e8f094090af879f8bc16ddce6db71e5ed745e7ed806a98e09",
		Target: "e2361a00",
	}

	share := <-stratumClient.Shares

	if share.Result != "03deef54ac208e5c4c41b608fa3c37436c5350858766d332fffbd8b06efc0700" {
		t.Errorf("Invalid share result '%v'", share.Result)
	}
}
