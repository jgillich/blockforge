package processor

import "gitlab.com/blockforge/blockforge/pkg/hwloc"

type Processor struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	// TODO
	// Cores         []Core `json:"cores"`
	PhysicalCores int `json:"physical_cores"`
	VirtualCores  int `json:"virtual_cores"`
}

func GetProcessors() ([]*Processor, error) {
	processors := []*Processor{}

	h, err := hwloc.NewTopology(hwloc.TopologyFlagThisSystem)
	if err != nil {
		return nil, err
	}

	for depth := uint(0); depth < uint(h.GetNbobjsByType(hwloc.ObjectTypePackage)); depth++ {
		cpuObj := h.GetObjByType(hwloc.ObjectTypePackage, depth)

		processor := Processor{
			Index:         int(depth),
			Name:          cpuObj.InfoByName("CPUModel"),
			PhysicalCores: h.GetNbobjsInsideCPUSetByType(cpuObj.CPUSet(), hwloc.ObjectTypeCore),
			VirtualCores:  h.GetNbobjsInsideCPUSetByType(cpuObj.CPUSet(), hwloc.ObjectTypePU),
		}

		processors = append(processors, &processor)
	}

	return processors, nil
}
