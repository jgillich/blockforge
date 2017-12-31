package hardware

import (
	"gitlab.com/jgillich/autominer/pkg/hwloc"
)

type Hardware struct {
	CPUs []CPU
	GPUs []GPU
}

type CPU struct {
	Model         string
	PhysicalCores int
	VirtualCores  int
}

type GPU struct {
	Model  string
	OpenCL bool
	CUDA   bool
}

func NewHardware() (*Hardware, error) {
	hw := Hardware{}

	_, err := hwloc.NewTopology(hwloc.TopologyFlagWholeSystem)
	if err != nil {
		return nil, err
	}

	return &hw, nil
}
