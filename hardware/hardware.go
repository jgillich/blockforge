package hardware

import (
	"strconv"
	"strings"

	"gitlab.com/jgillich/autominer/pkg/hwloc"
)

type GPUBackend string

var (
	OpenCLBackend GPUBackend = "OpenCL"
	CUDABackend   GPUBackend = "CUDA"
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
	Model    string
	Backend  GPUBackend
	Memory   int
	Platform int
	Index    int
}

type CUDA struct {
}

func NewHardware() (*Hardware, error) {
	hw := Hardware{}

	h, err := hwloc.NewTopology(hwloc.TopologyFlagWholeSystem)
	if err != nil {
		return nil, err
	}

	num := h.GetNbobjsByType(hwloc.ObjectTypeOsDevice)

	for i := 0; i < num; i++ {
		o := h.GetObjByType(hwloc.ObjectTypeOsDevice, i)

		gpu := GPU{
			Model:   o.InfoByName("GPUModel"),
			Backend: GPUBackend(o.InfoByName("Backend")),
		}

		if gpu.Backend == OpenCLBackend {
			name := o.Name()[:6]

			gpu.Memory, err = strconv.Atoi(o.InfoByName("OpenCLGlobalMemorySize"))
			if err != nil {
				continue
			}

			gpu.Platform, err = strconv.Atoi(name)
			if err != nil {
				continue
			}

			i := strings.Index(name, "d")
			gpu.Index, err = strconv.Atoi(name[:i])
			if err != nil {
				continue
			}
			gpu.Index++

			hw.GPUs = append(hw.GPUs, gpu)
		} else if gpu.Backend == CUDABackend {
			name := o.Name()[:4]

			gpu.Memory, err = strconv.Atoi(o.InfoByName("CUDAGlobalMemorySize"))
			if err != nil {
				continue
			}

			gpu.Index, err = strconv.Atoi(name)
			if err != nil {
				continue
			}

			hw.GPUs = append(hw.GPUs, gpu)
		} else {
			continue
		}

	}

	return &hw, nil
}
