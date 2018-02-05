package stratum

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"

	"go.uber.org/atomic"

	"github.com/Jeffail/gabs"

	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	clients[ProtocolJsonrpc] = newJsonrpcClient
}

type jsonrpcClient struct {
	conn    net.Conn
	jobs    chan Job
	minerId string
	pool    Pool
	closed  atomic.Bool
}

func newJsonrpcClient(pool Pool) (Client, error) {
	url, err := url.Parse(pool.URL)
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	switch url.Scheme {
	case "stratum+tls":
		certs, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		conn, err = tls.Dial("tcp", url.Host, &tls.Config{RootCAs: certs})
		if err != nil {
			return nil, err
		}
	case "stratum+tcp":
		conn, err = net.Dial("tcp", url.Host)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported stratum protocol '%v'", url.Scheme)
	}

	c := jsonrpcClient{
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
				if c.closed.Load() {
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

func (c *jsonrpcClient) Jobs() chan Job {
	return c.jobs
}

func (c *jsonrpcClient) SubmitShare(share *Share) {
	share.MinerId = c.minerId
	sendMessage(c.conn, &Message{
		Id:     1,
		Method: "submit",
		Params: share,
	})
	log.Info("share submitted")
}

func (c *jsonrpcClient) Close() error {
	c.closed.Store(true)
	close(c.jobs)
	return c.conn.Close()
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

	job := Job{
		JobId:  jobId,
		Blob:   blob,
		Target: target,
	}

	log.Debugf("got job '%+v'", jobId)

	c.jobs <- job
}
