package worker

type StratumTestClient struct {
	jobs   chan interface{}
	Shares chan interface{}
}

func NewStratumTestClient() *StratumTestClient {
	return &StratumTestClient{
		jobs:   make(chan interface{}, 10),
		Shares: make(chan interface{}),
	}
}

func (c *StratumTestClient) Close() error {
	return nil
}

func (c *StratumTestClient) GetJob() interface{} {
	j, ok := <-c.jobs
	if !ok {
		return nil
	}
	return j
}

func (c *StratumTestClient) SubmitShare(share interface{}) {
	c.Shares <- share
}
