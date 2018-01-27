package worker

import (
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
