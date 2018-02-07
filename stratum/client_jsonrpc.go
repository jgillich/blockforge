package stratum

import (
	"encoding/json"
	"io"

	"go.uber.org/atomic"

	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	clients[ProtocolJsonrpc] = newJsonrpcClient
}

type JsonrpcJob struct {
	JobId  string `json:"job_id"`
	Blob   string `json:"blob"`
	Target string `json:"target"`
}

type JsonrpcShare struct {
	MinerId string `json:"id"`
	JobId   string `json:"job_id"`
	Nonce   string `json:"nonce"`
	Result  string `json:"result"`
}

type jsonrpcClient struct {
	conn    *poolConn
	jobs    chan JsonrpcJob
	minerId string
	pool    Pool
	closed  atomic.Bool
}

type loginResult struct {
	Job     JsonrpcJob `json:"job"`
	Status  string     `json:"status"`
	MinerId string     `json:"id"`
}

func newJsonrpcClient(pool Pool) (Client, error) {
	conn, err := pool.dial()
	if err != nil {
		return nil, err
	}

	params, err := json.Marshal(map[string]string{
		"login": pool.User,
		"pass":  pool.Pass,
		"agent": agent,
	})
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "login",
		Params: params,
	}); err != nil {
		return nil, err
	}

	msg, err := conn.getMessage()
	if err != nil {
		return nil, err
	}

	var result loginResult
	if err := json.Unmarshal(msg.Result, &result); err != nil {
		return nil, err
	}

	c := &jsonrpcClient{
		minerId: result.MinerId,
		jobs:    make(chan JsonrpcJob, 10),
		pool:    pool,
		conn:    conn,
	}

	go c.loop()

	return c, nil
}

func (c *jsonrpcClient) loop() {
	for {
		msg, err := c.conn.getMessage()
		if err != nil {
			if c.closed.Load() {
				return
			}
			if err == io.EOF {
				log.Error("stratum server closed the connection, aborting")
				return
			}
			log.Error(err)
			continue
		}

		switch msg.Method {
		case "job":
			var job JsonrpcJob
			if err := json.Unmarshal(msg.Params, &job); err != nil {
				log.Errorw("error parsing job", "err", err)
				continue
			}
			c.jobs <- job
		}

	}
}

func (c *jsonrpcClient) GetJob() interface{} {
	j, ok := <-c.jobs
	if !ok {
		return nil
	}
	return j
}

func (c *jsonrpcClient) SubmitShare(in interface{}) {
	share := in.(*JsonrpcShare)
	share.MinerId = c.minerId

	params, err := json.Marshal(share)
	if err != nil {
		log.Errorw("error while serializing share", "err", err)
		return
	}

	c.conn.putMessage(&message{
		Id:     1,
		Method: "submit",
		Params: params,
	})
	log.Info("share submitted")
}

func (c *jsonrpcClient) Close() error {
	c.closed.Store(true)
	close(c.jobs)
	return c.conn.close()
}
