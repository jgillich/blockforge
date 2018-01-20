package stratum

import (
	"fmt"
	"net"

	"github.com/Jeffail/gabs"

	"gitlab.com/jgillich/autominer/log"
)

func init() {
	clients["jsonrpc"] = NewJsonrpcClient
}

type JsonrpcClient struct {
	conn    net.Conn
	jobs    chan Job
	minerId string
	pool    Pool
	closed  bool
}

func NewJsonrpcClient(pool Pool) (Client, error) {
	conn, err := net.Dial("tcp", pool.URL)
	if err != nil {
		return nil, err
	}

	c := JsonrpcClient{
		jobs: make(chan Job, 10),
		pool: pool,
		conn: conn,
	}

	login := &Message{
		Id:     1,
		Method: "login",
		Params: map[string]string{
			"login": c.pool.User,
			"pass":  c.pool.Pass,
			"agent": agent,
		},
	}

	if err := sendMessage(c.conn, login); err != nil {
		return nil, err
	}

	message, err := readMessage(c.conn)
	if err != nil {
		return nil, err
	}

	if message.Id != login.Id {
		return nil, fmt.Errorf("expected message id '%v' but got '%v'", login.Id, message.Id)
	}

	result, err := gabs.Consume(message.Result)
	if err != nil {
		return nil, err
	}

	if result.Exists("job") {
		container := result.Path("job")
		minerId, ok := result.Path("id").Data().(string)
		if !ok {
			panic("missing id")
		}
		c.minerId = minerId

		c.parseJob(container)
	}

	go func() {
		for {
			message, err := readMessage(c.conn)
			if err != nil {
				if c.closed {
					return
				}
				log.Error(err)
				continue
			}

			if message.Method == "job" {
				params, err := gabs.Consume(message.Params)
				if err != nil {
					log.Error(err)
					continue
				}
				c.parseJob(params)
			}

		}
	}()

	return &c, nil
}

func (c *JsonrpcClient) Jobs() chan Job {
	return c.jobs
}

func (c *JsonrpcClient) SubmitShare(share *Share) {
	sendMessage(c.conn, &Message{
		Id:     1,
		Method: "submit",
		Params: share,
	})
}

func (c *JsonrpcClient) Close() error {
	c.closed = true
	close(c.jobs)
	return c.conn.Close()
}

func (c *JsonrpcClient) parseJob(data *gabs.Container) {
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

	job := Job{
		MinerId: c.minerId,
		JobId:   jobId,
		Blob:    blob,
		Target:  target,
	}

	log.Debugf("got job '%+v'", jobId)

	go func() {
		c.jobs <- job
	}()

}
