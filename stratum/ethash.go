package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"

	"gitlab.com/blockforge/blockforge/algo"

	"go.uber.org/atomic"

	"gitlab.com/blockforge/blockforge/algo/ethash"
	"gitlab.com/blockforge/blockforge/log"
	"gitlab.com/blockforge/blockforge/worker"
)

var NicehashProtocolError = fmt.Errorf("EthereumStratum protocol error")

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
	conn, err := pool.dial()
	if err != nil {
		return nil, err
	}

	subscribeParams, err := json.Marshal([]string{agent, "EthereumStratum/1.0.0"})
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.subscribe",
		Params: subscribeParams,
	}); err != nil {
		return nil, err
	}

	msg, err := conn.getMessage()
	if err != nil {
		return nil, err
	}

	var result []json.RawMessage
	err = json.Unmarshal(msg.Result, &result)
	if err != nil {
		return nil, err
	}

	if len(result) != 2 {
		return nil, NicehashProtocolError
	}

	var first []string
	err = json.Unmarshal(result[0], &first)
	if err != nil {
		return nil, err
	}

	if len(first) != 3 {
		return nil, NicehashProtocolError
	}

	if first[2] != "EthereumStratum/1.0.0" {
		return nil, NicehashProtocolError
	}

	var extraNonce string
	err = json.Unmarshal(result[1], &extraNonce)
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.extranonce.subscribe",
	}); err != nil {
		return nil, err
	}

	authorizeParams, err := json.Marshal([]string{pool.User, pool.Pass})
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "mining.authorize",
		Params: authorizeParams,
	}); err != nil {
		return nil, err
	}

	msg, err = conn.getMessage()
	if err != nil {
		return nil, err
	}

	var loginSuccess bool
	err = json.Unmarshal(msg.Result, &loginSuccess)
	if err != nil {
		return nil, err
	}

	if !loginSuccess {
		return nil, fmt.Errorf("login failed")
	}

	stratum := &Ethash{
		work: make(chan *ethash.Work),
		pool: pool,
		conn: conn,
		// If pool does not set difficulty before first job, then miner can assume difficulty 1 was being set
		target: ethash.DiffToTarget(1),
	}

	stratum.setExtraNonce(extraNonce)

	go stratum.loop()

	return stratum, nil
}

func (stratum *Ethash) loop() {
	for {
		msg, err := stratum.conn.getMessage()
		if err != nil {
			if stratum.closed.Load() {
				return
			}
			if err == io.EOF {
				// TODO log error and reconnect
				log.Error("stratum server closed the connection, aborting")
				stratum.Close()
				return
			}
			log.Error(err)
			continue
		}

		switch msg.Method {
		case "mining.notify":
			var params []interface{}
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				log.Errorw("error while parsing job", "err", err)
				continue
			}

			if len(params) < 3 {
				log.Errorw("invalid job params length")
				continue
			}

			var job ethashJob
			var ok bool
			job.JobId, ok = params[0].(string)
			if !ok {
				log.Errorw("invalid job id")
				continue
			}
			job.SeedHash, ok = params[1].(string)
			if !ok {
				log.Errorw("invalid seed hash")
				continue
			}
			job.HeaderHash, ok = params[2].(string)
			if !ok {
				log.Errorw("invalid header hash")
				continue
			}
			job.CleanJobs, ok = params[3].(bool)
			if !ok {
				log.Errorw("invalid cleanjobs")
				continue
			}

			work, err := stratum.getWork(job)
			if err != nil {
				log.Error(err)
			}
			stratum.work <- work
		case "mining.set_difficulty":
			var params []float32
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				log.Errorw("error while parsing difficulty", "err", err)
				continue
			}

			if len(params) != 1 {
				log.Errorw("invalid set_difficulty params length")
				continue
			}

			stratum.target = ethash.DiffToTarget(params[0])
		case "mining.set_extranonce":
			var params []string
			err = json.Unmarshal(msg.Params, &params)
			if err != nil {
				log.Errorw("error while parsing extranonce", "err", err)
				continue
			}

			if len(params) != 1 {
				log.Errorw("invalid set_extranonce params length")
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
		log.Errorw("error while decoding extraNonce", "extraNonce", nonce)
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
		log.Errorw("error while serializing share", "err", err)
		return
	}
	log.Infof("submitting share %+v", params)
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
	if a != algo.Ethash {
		panic("invalid algorithm requested in ethash stratum")
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
