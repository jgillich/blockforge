package stratum

import (
	"fmt"
	"io"

	"go.uber.org/atomic"

	"github.com/Jeffail/gabs"

	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	clients[ProtocolJsonrpc] = newJsonrpcClient
}

type JsonrpcJob struct {
	JobId  string
	Blob   string
	Target string
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

func newJsonrpcClient(pool Pool) (Client, error) {
	conn, err := pool.dial()
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "login",
		Params: map[string]string{
			"login": pool.User,
			"pass":  pool.Pass,
			"agent": agent,
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

	c := &jsonrpcClient{
		jobs: make(chan JsonrpcJob, 10),
		pool: pool,
		conn: conn,
	}

	if result.Exists("job") {
		container := result.Path("job")
		minerId, ok := result.Path("id").Data().(string)
		if !ok {
			return nil, fmt.Errorf("server did not send miner id")
		}
		c.minerId = minerId

		c.parseJob(container)
	}

	go c.loop()

	return c, nil
}

func (c *jsonrpcClient) loop() {
	for {
		msg, err := c.conn.getMessage()
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

		switch msg.Method {
		case "job":
			params, err := gabs.Consume(msg.Params)
			if err != nil {
				log.Error(err)
				continue
			}
			c.parseJob(params)
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
	c.conn.putMessage(&message{
		Id:     1,
		Method: "submit",
		Params: share,
	})
	log.Info("share submitted")
}

func (c *jsonrpcClient) Close() error {
	c.closed.Store(true)
	close(c.jobs)
	return c.conn.close()
}

func (c *jsonrpcClient) parseJob(data *gabs.Container) {
	jobId, ok := data.Path("job_id").Data().(string)
	if !ok {
		log.Error("job_id missing or malformed")
		return
	}

	blob, ok := data.Path("blob").Data().(string)
	if !ok {
		log.Error("blob missing or malformed")
		return
	}

	target, ok := data.Path("target").Data().(string)
	if !ok {
		log.Error("target missing or malformed")
		return
	}

	job := JsonrpcJob{
		JobId:  jobId,
		Blob:   blob,
		Target: target,
	}

	c.jobs <- job
}
