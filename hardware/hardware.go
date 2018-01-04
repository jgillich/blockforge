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

type CUDA struct {
}

func NewHardware() (*Hardware, error) {
	hw := Hardware{}

	h, err := hwloc.NewTopology(hwloc.TopologyFlagWholeSystem)
	if err != nil {
		return nil, err
	}

	//  n = hwloc_get_nbobjs_by_type(topology, HWLOC_OBJ_OS_DEVICE);

	num := h.GetNbobjsByType(hwloc.ObjectTypeOsDevice)

	for i := 0; i < num; i++ {
		o := h.GetObjByType(hwloc.ObjectTypeOsDevice, i)

		backend := o.InfoByName("Backend")

		if backend == "CUDA" {

		} else if backend == "OpenCL" {

		}

		/*
		           assert(!strncmp(obj->name, "cuda", 4));
		           devid = atoi(obj->name + 4);
		           printf("CUDA device %d\n", devid);

		           s = hwloc_obj_get_info_by_name(obj, "GPUModel");
		           if (s)
		             printf("Model: %s\n", s);

		           s = hwloc_obj_get_info_by_name(obj, "CUDAGlobalMemorySize");
		           if (s)
		             printf("Memory: %s\n", s);

		           s = hwloc_obj_get_info_by_name(obj, "CUDAMultiProcessors");
		           if (s)
		           {
		             int mp = atoi(s);
		             s = hwloc_obj_get_info_by_name(obj, "CUDACoresPerMP");
		             if (s) {
		               int mp_cores = atoi(s);
		               printf("Cores: %d\n", mp * mp_cores);
		   					}
		*/
	}

	//  obj = hwloc_get_obj_by_type(topology, HWLOC_OBJ_OS_DEVICE, i);

	return &hw, nil
}
