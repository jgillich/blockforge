package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/Jeffail/gabs"
)

var agent = "coinstack/1.0.0"

type Message struct {
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Result  interface{} `json:"result"`
	Error   *Error      `json:"error"`
	Jsonrpc string      `json:"jsonrpc"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Job struct {
	MinerId string
	JobId   string
	Blob    string
	Target  string
}

type Share struct {
	MinerId string `json:"id"`
	JobId   string `json:"job_id"`
	Nonce   string `json:"nonce"`
	Result  string `json:"result"`
}

type Client struct {
	Jobs    chan Job
	conn    net.Conn
	pool    Pool
	id      int
	minerId string
}

func NewClient(pool Pool) (*Client, error) {
	if pool.Protocol == "" {
		pool.Protocol = ProtocolStandard
	}

	conn, err := net.Dial("tcp", pool.URL)
	if err != nil {
		return nil, err
	}
	log.Printf("connected to %v", pool.URL)
	return &Client{
		Jobs: make(chan Job, 10),
		id:   1,
		pool: pool,
		conn: conn,
	}, nil
}

func (c *Client) SubmitShare(share *Share) {
	c.send(&Message{
		Method: "submit",
		Params: share,
	})
}

func (c *Client) Connect() error {

	if err := c.login(); err != nil {
		return err
	}

	go func() {
		for {
			message, err := c.read()
			if err != nil {
				panic(err)
			}

			if message.Method == "job" {
				params, err := gabs.Consume(message.Params)
				if err != nil {
					panic(err)
				}
				c.parseJob(params)
			}

		}
	}()

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) send(message *Message) error {

	if message.Id == 0 {
		message.Id = c.id
		c.id = c.id + 1
	}

	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	log.Printf("sending message %v", string(msg))

	_, err = fmt.Fprintf(c.conn, "%v\n", string(msg))
	return err
}

func (c *Client) read() (*Message, error) {
	s, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		return nil, err
	}

	log.Printf("received message %v", string(s))

	var message Message
	err = json.Unmarshal([]byte(s), &message)
	if err != nil {
		return nil, err
	}

	if message.Error != nil {
		return nil, fmt.Errorf("server responded with error '%v': '%v'", message.Error.Code, message.Error.Message)
	}

	return &message, nil
}

func (c *Client) login() error {

	if c.pool.Protocol == ProtocolNicehash {

		subscribe := &Message{
			Method: "mining.subscribe",
			Params: []string{
				agent,
				"EthereumStratum/1.0.0",
			},
		}

		if err := c.send(subscribe); err != nil {
			return err
		}

		message, err := c.read()
		if err != nil {
			return err
		}

		if message.Id != subscribe.Id {
			return fmt.Errorf("expected message id '%v' but got '%v'", subscribe.Id, message.Id)
		}

		// TODO validate message

		authorize := &Message{
			Method: "mining.authorize",
			Params: []string{c.pool.User, c.pool.Pass},
		}

		if err := c.send(authorize); err != nil {
			return err
		}

		message, err = c.read()
		if err != nil {
			return err
		}

		if message.Id != authorize.Id {
			return fmt.Errorf("expected message id '%v' but got '%v'", authorize.Id, message.Id)
		}
		// TODO check result: true

	} else {

		login := &Message{
			Method: "login",
			Params: map[string]string{
				"login": c.pool.User,
				"pass":  c.pool.Pass,
				"agent": agent,
			},
		}

		if err := c.send(login); err != nil {
			return err
		}

		message, err := c.read()
		if err != nil {
			return err
		}

		if message.Id != login.Id {
			return fmt.Errorf("expected message id '%v' but got '%v'", login.Id, message.Id)
		}

		result, err := gabs.Consume(message.Result)
		if err != nil {
			return err
		}

		log.Print("authenticated")

		if result.Exists("job") {
			container := result.Path("job")
			minerId, ok := result.Path("id").Data().(string)
			if !ok {
				panic("missing id")
			}
			c.minerId = minerId

			c.parseJob(container)
		}
	}
	return nil
}

func (c *Client) parseJob(data *gabs.Container) {
	jobId, ok := data.Path("job_id").Data().(string)
	if !ok {
		log.Printf("job_id not ok")
		return
	}

	blob, ok := data.Path("blob").Data().(string)
	if !ok {
		log.Printf("blob not ok")
		return
	}

	target, ok := data.Path("target").Data().(string)
	if !ok {
		log.Printf("target not ok")
		return
	}

	job := Job{
		MinerId: c.minerId,
		JobId:   jobId,
		Blob:    blob,
		Target:  target,
	}

	log.Printf("got job '%+v'", jobId)

	go func() {
		c.Jobs <- job
	}()

}
