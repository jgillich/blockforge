package stratum

import (
	"fmt"
	"io"

	"go.uber.org/atomic"

	"github.com/Jeffail/gabs"

	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	clients[ProtocolNicehash] = newNicehashClient
}

type NicehashJob struct {
	Difficulty float64
	SeedHash   string
	HeaderHash string
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
	difficulty float64
}

func newNicehashClient(pool Pool) (Client, error) {
	conn, err := pool.dial()
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.subscribe",
		Params: []string{
			agent,
			"EthereumStratum/1.0.0",
		},
	}); err != nil {
		return nil, err
	}

	msg, err := conn.getMessage()
	if err != nil {
		return nil, err
	}

	result, err := gabs.Consume(msg.Result)
	if err != nil {
		return nil, err
	}

	results, err := result.Children()
	if err != nil {
		return nil, err
	}

	inner, err := results[0].Children()
	if err != nil {
		return nil, err
	}

	if inner[2].Data() != "EthereumStratum/1.0.0" {
		return nil, fmt.Errorf("stratum server does not use protocol EthereumStratum/1.0.0")
	}

	extraNonce, ok := results[1].Data().(string)
	if !ok || extraNonce == "" {
		return nil, fmt.Errorf("missing or invalid etraNonce: '%v'", results[1].String())
	}

	/*
		if err := sendMessage(conn, &message{
			Id:     1,
			Method: "mining.extranonce.subscribe",
		}); err != nil {
			return nil, err
		}
	*/

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.authorize",
		Params: []string{pool.User, pool.Pass},
	}); err != nil {
		return nil, err
	}

	msg, err = conn.getMessage()
	if err != nil {
		return nil, err
	}

	result, err = gabs.Consume(msg.Result)
	if err != nil {
		return nil, err
	}

	r, ok := result.Data().(bool)
	if !r || !ok {
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
		message, err := c.conn.getMessage()
		if err != nil {
			if err == io.EOF {
				if c.closed.Load() {
					return
				}
				log.Error("stratum server closed the connection, aborting")
				c.Close()
				return
			}
			log.Error(err)
			continue
		}

		params, err := gabs.Consume(message.Params)
		if err != nil {
			log.Error(err)
			continue
		}

		switch message.Method {
		case "mining.notify":
			children, err := params.Children()
			if err != nil {
				log.Error(err)
				continue
			}
			seedHash, ok := children[0].Data().(string)
			if !ok {
				log.Error("received invalid seedHash")
				continue
			}
			headerHash, ok := children[1].Data().(string)
			if !ok {
				log.Error("received invalid headerHash")
				continue
			}

			c.jobs <- NicehashJob{
				Difficulty: c.difficulty,
				SeedHash:   seedHash,
				HeaderHash: headerHash,
			}

		case "mining.set_difficulty":
			children, err := params.Children()
			if err != nil {
				log.Error(err)
				continue
			}

			difficulty, ok := children[0].Data().(float64)
			if !ok {
				log.Error("received invalid difficulty")
				continue
			}

			c.difficulty = difficulty
		case "mining.set_extranonce":
			children, err := params.Children()
			if err != nil {
				log.Error(err)
				continue
			}

			extraNonce, ok := children[0].Data().(string)
			if !ok {
				log.Error("received invalid extraNonce")
				continue
			}

			c.extraNonce = extraNonce
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
