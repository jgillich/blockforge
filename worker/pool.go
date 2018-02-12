package worker

type Pool struct {
	Results  chan interface{}
	channels []chan interface{}
}

func (pool *Pool) Spawn(f func(chan interface{}, chan interface{})) {
	ch := make(chan interface{}, 1)
	pool.channels = append(pool.channels, ch)
	go f(ch, pool.Results)
}

func (pool *Pool) Dispatch(work interface{}) {
	for _, ch := range pool.channels {
		ch <- work
	}
}

func (pool *Pool) Close() {
	for _, ch := range pool.channels {
		close(ch)
	}
}
