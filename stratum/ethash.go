package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/getsentry/raven-go"

	"gitlab.com/blockforge/blockforge/algo"

	"go.uber.org/atomic"

	"gitlab.com/blockforge/blockforge/algo/ethash"
	"gitlab.com/blockforge/blockforge/log"
	"gitlab.com/blockforge/blockforge/worker"
)

func init() {
	clients[ProtocolEthereum] = NewEthash
}

type ethashJob struct {
	JobId      string
	SeedHash   string
	HeaderHash string
	CleanJobs  bool
}

type ethashShare struct {
	JobId string `json:"id"`
	Nonce string `json:"nonce"`
}

type Ethash struct {
	work chan *ethash.Work
	conn *poolConn

	minerId    string
	pool       Pool
	closed     atomic.Bool
	extraNonce uint64
	target     *big.Int
}

type subscribeResult struct {
}

func NewEthash(pool Pool) (Client, error) {
	stratum := &Ethash{
		// buffered so we can keep parsing messages and avoid panic when stop is called during dag generation
		work: make(chan *ethash.Work, 25),
		pool: pool,
	}

	if err := stratum.login(); err != nil {
		return nil, err
	}

	go func() {
		for {
			stratum.loop()
			if stratum.closed.Load() {
				return
			}
			for err := stratum.login(); err != nil; err = stratum.login() {
				log.Error(err)
				log.Info("failed to connect to stratum server, sleeping for 10 seconds...")
				time.Sleep(time.Second * 10)
			}
		}
	}()

	return stratum, nil
}

func (stratum *Ethash) login() error {
	conn, err := stratum.pool.dial()
	if err != nil {
		return err
	}

	subscribeParams, err := json.Marshal([]string{agent, "EthereumStratum/1.0.0"})
	if err != nil {
		return err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.subscribe",
		Params: subscribeParams,
	}); err != nil {
		return err
	}

	msg, err := conn.getMessage()
	if err != nil {
		return err
	}

	var result []json.RawMessage
	err = json.Unmarshal(msg.Result, &result)
	if err != nil {
		return err
	}

	if len(result) != 2 {
		return ProtocolError
	}

	var first []string
	err = json.Unmarshal(result[0], &first)
	if err != nil {
		return err
	}

	if len(first) != 3 {
		return ProtocolError
	}

	if first[2] != "EthereumStratum/1.0.0" {
		return ProtocolError
	}

	var extraNonce string
	err = json.Unmarshal(result[1], &extraNonce)
	if err != nil {
		return err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.extranonce.subscribe",
	}); err != nil {
		return err
	}

	authorizeParams, err := json.Marshal([]string{stratum.pool.User, stratum.pool.Pass})
	if err != nil {
		return err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.authorize",
		Params: authorizeParams,
	}); err != nil {
		return err
	}

	msg, err = conn.getMessage()
	if err != nil {
		return err
	}

	var loginSuccess bool
	err = json.Unmarshal(msg.Result, &loginSuccess)
	if err != nil {
		return err
	}

	if !loginSuccess {
		return fmt.Errorf("login failed")
	}

	// If pool does not set difficulty before first job, then miner can assume difficulty 1 was being set
	stratum.target = ethash.DiffToTarget(1)

	stratum.conn = conn

	stratum.setExtraNonce(extraNonce)

	return nil
}

func (stratum *Ethash) loop() {
	for {
		msg, err := stratum.conn.getMessage()
		if err != nil {
			if stratum.closed.Load() || err == io.EOF {
				return
			}
			stratum.protoErr(err)
			continue
		}

		switch msg.Method {
		case "mining.notify":
			var params []interface{}
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				stratum.protoErr(err)
				continue
			}

			if len(params) < 3 {
				stratum.protoErr(ProtocolError)
				continue
			}

			var job ethashJob
			var ok bool
			job.JobId, ok = params[0].(string)
			if !ok {
				stratum.protoErr(ProtocolError)
				continue
			}
			job.SeedHash, ok = params[1].(string)
			if !ok {
				stratum.protoErr(ProtocolError)
				continue
			}
			job.HeaderHash, ok = params[2].(string)
			if !ok {
				stratum.protoErr(ProtocolError)
				continue
			}
			job.CleanJobs, ok = params[3].(bool)
			if !ok {
				stratum.protoErr(ProtocolError)
				continue
			}

			work, err := stratum.getWork(job)
			if err != nil {
				stratum.protoErr(err)
				continue
			}
			stratum.work <- work
		case "mining.set_difficulty":
			var params []float32
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				stratum.protoErr(err)
				continue
			}

			if len(params) != 1 {
				stratum.protoErr(ProtocolError)
				continue
			}

			log.Infof("job difficulty %v", params[0])

			stratum.target = ethash.DiffToTarget(params[0])
		case "mining.set_extranonce":
			var params []string
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				stratum.protoErr(err)
				continue
			}

			if len(params) != 1 {
				stratum.protoErr(ProtocolError)
				continue
			}

			stratum.setExtraNonce(params[0])
		case "client.get_version":
			// TODO
		}
	}
}

func (stratum *Ethash) getWork(job ethashJob) (*ethash.Work, error) {
	header, err := hex.DecodeString(strings.TrimPrefix(job.SeedHash, "0x"))
	if err != nil {
		return nil, err
	}

	return &ethash.Work{
		JobId:      job.JobId,
		Seedhash:   job.SeedHash,
		Header:     header,
		Target:     stratum.target,
		ExtraNonce: stratum.extraNonce,
	}, nil

}

func (stratum *Ethash) setExtraNonce(nonce string) {
	for i := len(nonce); i < 16; i++ {
		nonce += "0"
	}
	extraNonceBytes, err := hex.DecodeString(nonce)
	if err != nil {
		stratum.protoErr(err)
		return
	}
	stratum.extraNonce = binary.BigEndian.Uint64(extraNonceBytes)
}

func (stratum *Ethash) submit(share ethash.Share) {
	params, err := json.Marshal([]string{
		stratum.pool.User,
		share.JobId,
		fmt.Sprintf("%v", share.Nonce),
	})
	if err != nil {
		stratum.protoErr(err)
		return
	}
	log.Infof("submitting share %v", string(params))
	stratum.conn.putMessage(&message{
		Id:     1,
		Method: "mining.submit",
		Params: params,
	})
	log.Info("share submitted")
}

func (stratum *Ethash) Close() error {
	stratum.closed.Store(true)
	close(stratum.work)
	return stratum.conn.close()
}

func (stratum *Ethash) Worker(a algo.Algo) worker.Worker {
	_, ok := a.(*ethash.Algo)
	if !ok {
		log.Panic("invalid algorithm requested in ethash stratum")
	}

	shares := make(chan ethash.Share, 1)
	go func() {
		defer close(shares)
		for share := range shares {
			if stratum.closed.Load() {
				break
			}
			stratum.submit(share)
		}
	}()

	return &worker.Ethash{
		Work:   stratum.work,
		Shares: shares,
	}
}

func (stratum *Ethash) protoErr(err error) {
	raven.CaptureError(err, map[string]string{
		"url": stratum.pool.URL,
	})
	log.Error(err)
}
