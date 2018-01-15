package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Client struct {
	conn net.Conn
	pool Pool
	id   int
}

type Message struct {
	Id     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Result interface{} `json:"result"`
	Error  *Error      `json:"error"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(pool Pool) (*Client, error) {
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

	if err := c.subscribe(); err != nil {
		return err
	}

	if err := c.authorize(); err != nil {
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

func (c *Client) subscribe() error {
	subscribe := &Message{
		Id:     1,
		Method: "mining.subscribe",
		Params: []string{
			"coinstack/1.0.0",
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

	return nil
}

func (c *Client) authorize() error {

	authorize := &Message{
		Id:     2,
		Method: "mining.authorize",
		Params: []string{c.pool.User, c.pool.Pass},
	}

	if err := c.send(authorize); err != nil {
		return err
	}

	message, err := c.read()
	if err != nil {
		return err
	}

	if message.Id != authorize.Id {
		return fmt.Errorf("expected message id '%v' but got '%v'", authorize.Id, message.Id)
	}

	// TODO check result: true

	return nil
}
