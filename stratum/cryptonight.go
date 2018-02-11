package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

	"gitlab.com/blockforge/blockforge/algo"
	"gitlab.com/blockforge/blockforge/algo/cryptonight"
	"gitlab.com/blockforge/blockforge/worker"

	"go.uber.org/atomic"

	"gitlab.com/blockforge/blockforge/log"
)

func init() {
	clients[ProtocolCryptonight] = NewCryptonight
}

type cryptonightJob struct {
	JobId  string `json:"job_id"`
	Blob   string `json:"blob"`
	Target string `json:"target"`
}

type cryptonightShare struct {
	MinerId string `json:"id"`
	JobId   string `json:"job_id"`
	Nonce   string `json:"nonce"`
	Result  string `json:"result"`
}

type cryptonightLoginResult struct {
	Job     cryptonightJob `json:"job"`
	Status  string         `json:"status"`
	MinerId string         `json:"id"`
}

type Cryptonight struct {
	work    chan *cryptonight.Work
	conn    *poolConn
	minerId string
	pool    Pool
	closed  atomic.Bool
	jobId   atomic.String
}

func NewCryptonight(pool Pool) (Client, error) {
	conn, err := pool.dial()
	if err != nil {
		return nil, err
	}

	params, err := json.Marshal(map[string]string{
		"login": pool.User,
		"pass":  pool.Pass,
		"agent": agent,
	})
	if err != nil {
		return nil, err
	}

	if err := conn.putMessage(&message{
		Id:     1,
		Method: "login",
		Params: params,
	}); err != nil {
		return nil, err
	}

	msg, err := conn.getMessage()
	if err != nil {
		return nil, err
	}

	var result cryptonightLoginResult
	if err := json.Unmarshal(msg.Result, &result); err != nil {
		return nil, err
	}

	stratum := &Cryptonight{
		work:    make(chan *cryptonight.Work, 1),
		minerId: result.MinerId,
		pool:    pool,
		conn:    conn,
	}

	work, err := stratum.getWork(result.Job)
	if err != nil {
		return nil, err
	}
	stratum.work <- work

	go stratum.loop()

	return stratum, nil
}

func (stratum *Cryptonight) loop() {
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
		case "job":
			var job cryptonightJob
			if err := json.Unmarshal(msg.Params, &job); err != nil {
				log.Errorw("error parsing job", "err", err)
				continue
			}
			work, err := stratum.getWork(job)
			if err != nil {
				log.Error(err)
				continue
			}
			stratum.work <- work
		}

	}
}

func (stratum *Cryptonight) getWork(job cryptonightJob) (*cryptonight.Work, error) {
	input, err := hex.DecodeString(job.Blob)
	if err != nil {
		log.Errorw("malformed blob", "job", job)
		return nil, errors.New("malformed blob")
	}

	t, err := hex.DecodeString(job.Target)
	if err != nil {
		return nil, errors.New("malformed target")
	}

	var target uint64
	switch len(job.Target) {
	case 8:
		t32 := uint64(binary.LittleEndian.Uint32(t))
		target = math.MaxUint64 / (math.MaxUint32 / t32)
	case 16:
		target = binary.LittleEndian.Uint64(t)
	default:
		return nil, errors.New("unsupported target length")
	}

	log.Infof("job difficulty %v", math.MaxUint64/target)

	stratum.jobId.Store(job.JobId)

	return &cryptonight.Work{
		JobId:  job.JobId,
		Input:  input,
		Target: target,
	}, nil
}

func (stratum *Cryptonight) submit(in cryptonight.Share) {
	nonce := make([]byte, 4)
	binary.LittleEndian.PutUint32(nonce, in.Nonce)

	share := cryptonightShare{
		JobId:   in.JobId,
		Result:  fmt.Sprintf("%x", in.Result),
		Nonce:   fmt.Sprintf("%08x", nonce),
		MinerId: stratum.minerId,
	}

	params, err := json.Marshal(share)
	if err != nil {
		log.Errorw("error while serializing share", "err", err)
		return
	}

	stratum.conn.putMessage(&message{
		Id:     1,
		Method: "submit",
		Params: params,
	})
	log.Info("share submitted")
}

func (stratum *Cryptonight) Close() error {
	stratum.closed.Store(true)
	close(stratum.work)
	return stratum.conn.close()
}

func (stratum *Cryptonight) Worker(a algo.Algo) worker.Worker {
	var lite bool
	switch a {
	case algo.Cryptonight:
		lite = false
	case algo.CryptonightLite:
		lite = true
	default:
		panic("invalid algorithm requested in cryptonight stratum")
	}

	shares := make(chan cryptonight.Share, 1)
	go func() {
		defer close(shares)
		for share := range shares {
			if stratum.closed.Load() {
				break
			}
			if stratum.jobId.Load() != share.JobId {
				log.Info("share skipped")
				continue
			}
			stratum.submit(share)
		}
	}()

	return &worker.Cryptonight{
		Lite:   lite,
		Work:   stratum.work,
		Shares: shares,
	}
}
