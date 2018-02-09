package stratum

import (
	"encoding/json"
	"fmt"
	"io"

	"go.uber.org/atomic"

	"gitlab.com/blockforge/blockforge/log"
)

var NicehashProtocolError = fmt.Errorf("EthereumStratum protocol error")

func init() {
	clients[ProtocolNicehash] = newNicehashClient
}

type NicehashJob struct {
	JobId      string
	Difficulty float32
	SeedHash   string
	HeaderHash string
	CleanJobs  bool
}

type NicehashShare struct {
	MinerId string `json:"id"`
	JobId   string `json:"job_id"`
	Nonce   string `json:"nonce"`
	Result  string `json:"result"`
}

type nicehashClient struct {
	conn       *poolConn
	jobs       chan NicehashJob
	minerId    string
	pool       Pool
	closed     atomic.Bool
	extraNonce string
	difficulty float32
}

type subscribeResult struct {
}

func newNicehashClient(pool Pool) (Client, error) {
	conn, err := pool.dial()
	if err != nil {
		return nil, err
	}

	subscribeParams, err := json.Marshal([]string{agent, "EthereumStratum/1.0.0"})
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.subscribe",
		Params: subscribeParams,
	}); err != nil {
		return nil, err
	}

	msg, err := conn.getMessage()
	if err != nil {
		return nil, err
	}

	var result []json.RawMessage
	err = json.Unmarshal(msg.Result, &result)
	if err != nil {
		return nil, err
	}

	if len(result) != 2 {
		return nil, NicehashProtocolError
	}

	var first []string
	err = json.Unmarshal(result[0], &first)
	if err != nil {
		return nil, err
	}

	if len(first) != 3 {
		return nil, NicehashProtocolError
	}

	if first[2] != "EthereumStratum/1.0.0" {
		return nil, NicehashProtocolError
	}

	var extraNonce string
	err = json.Unmarshal(result[1], &extraNonce)
	if err != nil {
		return nil, err
	}

	/*
		if err := sendMessage(conn, &message{
			Id:     1,
			Method: "mining.extranonce.subscribe",
		}); err != nil {
			return nil, err
		}
	*/

	authorizeParams, err := json.Marshal([]string{pool.User, pool.Pass})
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.authorize",
		Params: authorizeParams,
	}); err != nil {
		return nil, err
	}

	msg, err = conn.getMessage()
	if err != nil {
		return nil, err
	}

	var loginSuccess bool
	err = json.Unmarshal(msg.Result, &loginSuccess)
	if err != nil {
		return nil, err
	}

	if !loginSuccess {
		return nil, fmt.Errorf("login failed")
	}

	c := &nicehashClient{
		jobs:       make(chan NicehashJob, 10),
		pool:       pool,
		conn:       conn,
		extraNonce: extraNonce,
		// If pool does not set difficulty before first job, then miner can assume difficulty 1 was being set
		difficulty: 1,
	}

	go c.loop()

	return c, nil
}

func (c *nicehashClient) loop() {
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
		case "mining.notify":

			var params []interface{}
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				log.Errorw(NicehashProtocolError.Error(), "err", err)
				continue
			}

			if len(params) < 3 {
				log.Errorw("invalid job params length")
				continue
			}

			job := NicehashJob{}
			var ok bool
			job.JobId, ok = params[0].(string)
			if !ok {
				log.Errorw("invalid job id")
				continue
			}
			job.SeedHash, ok = params[1].(string)
			if !ok {
				log.Errorw("invalid seed hash")
				continue
			}
			job.HeaderHash, ok = params[2].(string)
			if !ok {
				log.Errorw("invalid header hash")
				continue
			}
			job.CleanJobs, ok = params[3].(bool)
			if !ok {
				log.Errorw("invalid cleanjobs")
				continue
			}

			c.jobs <- job

		case "mining.set_difficulty":
			var params []float32
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				log.Errorw(NicehashProtocolError.Error(), "err", err)
				continue
			}

			if len(params) != 1 {
				log.Errorw("invalid set_difficulty params length")
				continue
			}

			c.difficulty = params[0]
		case "mining.set_extranonce":
			var params []string
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				log.Errorw(NicehashProtocolError.Error(), "err", err)
				continue
			}

			if len(params) != 1 {
				log.Errorw("invalid set_extranonce params length")
				continue
			}

			c.extraNonce = params[0]
		case "client.get_version":
		}
	}
}

func (c *nicehashClient) GetJob() interface{} {
	j, ok := <-c.jobs
	if !ok {
		return nil
	}
	return j
}

func (c *nicehashClient) SubmitShare(share interface{}) {

}

func (c *nicehashClient) Close() error {
	c.closed.Store(true)
	close(c.jobs)
	return c.conn.close()
}
