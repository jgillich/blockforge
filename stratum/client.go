package stratum

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"gitlab.com/blockforge/blockforge/log"
)

var clients = map[Protocol]clientFactory{}

type clientFactory func(Pool) (Client, error)

// NewClient creates a new client for a specified protocol
func NewClient(protocol Protocol, pool Pool) (Client, error) {
	factory, ok := clients[protocol]
	if !ok {
		return nil, fmt.Errorf("client for protocol '%v' does not exist", protocol)
	}

	return factory(pool)
}

// Client implements a variant of the Stratum protocol
type Client interface {
	// Close closes the stratum connetion
	Close() error
	// GetJob blocks until a new job is available, or returns nil when the connection has been closed
	GetJob() interface{}
	// SubmitShare submits a share to the server
	SubmitShare(interface{})
}

var agent = "blockforge/1.0.0"

type message struct {
	Id      int             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *stratumError   `json:"error,omitempty"`
	Jsonrpc string          `json:"jsonrpc,omitempty"`
}

type stratumError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type poolConn struct {
	conn net.Conn
}

func (p Pool) dial() (*poolConn, error) {
	url, err := url.Parse(p.URL)
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

	return &poolConn{conn}, nil
}

func (p *poolConn) putMessage(message *message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	log.Debugf("putMessage %v", string(msg))

	_, err = fmt.Fprintf(p.conn, "%v\n", string(msg))
	return err
}

func (p *poolConn) getMessage() (*message, error) {
	s, err := bufio.NewReader(p.conn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	log.Debugf("getMessage %v", strings.TrimRight(string(s), "\n"))

	var msg message
	err = json.Unmarshal([]byte(s), &msg)
	if err != nil {
		return nil, err
	}

	if msg.Error != nil {
		return nil, fmt.Errorf("server responded with error '%v': '%v'", msg.Error.Code, msg.Error.Message)
	}

	return &msg, nil
}

func (p *poolConn) close() error {
	return p.conn.Close()
}
