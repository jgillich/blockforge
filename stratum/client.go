package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

var agent = "coinstack/1.0.0"

type Client struct {
	conn net.Conn
	pool Pool
	id   int
}

type Message struct {
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Result  interface{} `json:"result"`
	Error   *Error      `json:"error"`
	Status  string      `json:"status"`
	Jsonrpc string      `json:"jsonrpc"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
		pool: pool,
		conn: conn,
	}, nil
}

func (c Client) Connect() error {

	if err := c.login(); err != nil {
		return err
	}

	for {
		_, err := c.read()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) send(message *Message) error {
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
			Id:     1,
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
			Id:     2,
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
			Id:     2,
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

	}

	return nil
}
