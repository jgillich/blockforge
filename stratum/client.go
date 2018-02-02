package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"gitlab.com/blockforge/blockforge/log"
)

var clients = map[Protocol]clientFactory{}

type clientFactory func(Pool) (Client, error)

func NewClient(protocol Protocol, pool Pool) (Client, error) {
	factory, ok := clients[protocol]
	if !ok {
		return nil, fmt.Errorf("client for protocol '%v' does not exist", protocol)
	}

	return factory(pool)
}

type Client interface {
	Close() error
	Jobs() chan Job
	SubmitShare(*Share)
}

var agent = "blockforge/1.0.0"

type Message struct {
	Id      int         `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	Jsonrpc string      `json:"jsonrpc,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Job struct {
	JobId  string
	Blob   string
	Target string
}

type Share struct {
	MinerId string `json:"id"`
	JobId   string `json:"job_id"`
	Nonce   string `json:"nonce"`
	Result  string `json:"result"`
}

func sendMessage(conn net.Conn, message *Message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	log.Debugf("sending message %v", string(msg))

	_, err = fmt.Fprintf(conn, "%v\n", string(msg))
	return err
}

func readMessage(conn net.Conn) (*Message, error) {
	s, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, err
	}

	log.Debugf("received message %v", strings.TrimRight(string(s), "\n"))

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

/*
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
*/
