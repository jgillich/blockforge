package worker

func init() {
	for _, c := range []string{"ETH", "ETC"} {
		workers[c] = newEthash
	}
}

type ethash struct {
	config Config
}

func newEthash(config Config) Worker {
	return &ethash{config}
}

func (w *ethash) Work() error {
	for {
		job := w.config.Stratum.GetJob()
		if job == nil {
			return nil
		}
	}
}

func (w *ethash) Stats() Stats {
	stats := Stats{
		CPUStats: []CPUStats{},
		GPUStats: []GPUStats{},
	}

	return stats
}

func (w *ethash) Capabilities() Capabilities {
	return Capabilities{
		CPU:    true,
		OpenCL: false,
		CUDA:   false,
	}
}
