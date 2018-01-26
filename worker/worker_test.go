package worker

import (
	"fmt"
	"testing"

	"gitlab.com/blockforge/blockforge/stratum"
)

type StratumTestClient struct {
	jobs   chan stratum.Job
	Shares chan stratum.Share
}

func NewStratumTestClient() *StratumTestClient {
	return &StratumTestClient{
		jobs:   make(chan stratum.Job, 10),
		Shares: make(chan stratum.Share),
	}
}

func (c *StratumTestClient) Close() error {
	return nil
}

func (c *StratumTestClient) Jobs() chan stratum.Job {
	return c.jobs
}

func (c *StratumTestClient) SubmitShare(share *stratum.Share) {
	c.Shares <- *share
}

func TestNonce(t *testing.T) {
	nonce := "0000001b"

	num := hexUint32([]byte(nonce))
	fmt.Println(num)
	formatted := fmtNonce(num)

	if nonce != formatted {
		t.Errorf("Failed to parse/format nonce, expected '%v' got '%v'", nonce, formatted)
	}

}

func TestHexUint64LE(t *testing.T) {
	res := hexUint64LE([]byte("3fa12800"))
	if res != 2662719 {
		t.Errorf("wrong uint expected '%v' got '%v'", 2662719, res)
	}
}
