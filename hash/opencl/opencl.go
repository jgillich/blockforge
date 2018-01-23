package opencl

import (
	"fmt"

	"github.com/jgillich/go-opencl/cl"
)

type opencl struct{}

func New() *opencl {
	o := &opencl{}

	platforms, err := cl.GetPlatforms()
	if err != nil {
		panic(err)
	}

	for _, p := range platforms {
		fmt.Printf("%v (%v)\n", p.Name(), p.Version())

		devices, err := p.GetDevices(cl.DeviceTypeGPU)
		if err != nil {
			panic(err)
		}

		for _, d := range devices {
			fmt.Println("  " + d.Name())
		}

	}

	return o
}
