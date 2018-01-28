package opencl

import (
	"github.com/jgillich/go-opencl/cl"
	"gitlab.com/blockforge/blockforge/log"
)

type Device struct {
	ptr      *cl.Device
	Platform *Platform `json:"platform"`
	Index    int       `json:"index"`
	Name     string    `json:"name"`
	Vendor   string    `json:"vendor"`
	Version  string    `json:"version"`
}

func GetDevices(platform *Platform) ([]*Device, error) {
	devices := []*Device{}

	clDevices, err := platform.ptr.GetDevices(cl.DeviceTypeGPU)
	if err != nil {
		return nil, err
	}

	for i, d := range clDevices {
		device := Device{
			ptr:      d,
			Platform: platform,
			Index:    i,
			Name:     d.Name(),
			Vendor:   d.Vendor(),
			Version:  d.Version(),
		}

		if !d.Available() {
			log.Debugw("skipping unavailable opencl device", "device", d.Name())
			continue
		}
		devices = append(devices, &device)
	}

	return devices, nil
}

func (d *Device) CL() *cl.Device {
	return d.ptr
}
