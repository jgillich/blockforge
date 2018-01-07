package coin

type Backend string

var (
	BackendCPU    Backend = "cpu"
	BackendOpenCL Backend = "opencl"
	BackendCUDA   Backend = "cuda"
)

type Miner interface {
	Start() error
	Stop()
	Stats() MinerStats
}

type MinerStats struct {
	Coin     string
	Hashrate float32
}

type MinerConfig struct {
	Coin       string
	Donate     int
	PoolURL    string
	PoolUser   string
	PoolPass   string
	PoolEmail  string
	Threads    int
	GPUIndexes []int
}
