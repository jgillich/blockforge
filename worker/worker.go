package worker

import (
	"gitlab.com/jgillich/autominer/stratum"
)

type Worker interface {
	Work(stratum.Job) (*stratum.Share, error)
}
