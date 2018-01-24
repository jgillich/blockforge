package opencl

import (
	"github.com/jgillich/go-opencl/cl"
	"gitlab.com/jgillich/autominer/log"
)

type Platform struct {
	ptr     *cl.Platform
	Index   int       `json:"index"`
	Name    string    `json:"name"`
	Version string    `json:"version"`
	Vendor  string    `json:"vendor"`
	Devices []*Device `json:"devices"`
}

func GetPlatforms() ([]*Platform, error) {
	platforms := []*Platform{}

	clPlatforms, err := cl.GetPlatforms()
	if err != nil {
		return nil, err
	}

	for i, p := range clPlatforms {
		platform := Platform{
			Index:   i,
			ptr:     p,
			Name:    p.Name(),
			Vendor:  p.Vendor(),
			Version: p.Version(),
		}

		platform.Devices, err = GetDevices(&platform)

		if err == cl.ErrDeviceNotFound {
			log.Debugw("skipping opencl platform without gpu devices", "platform", p.Name())
			continue
		}

		platforms = append(platforms, &platform)
	}

	return platforms, nil
}

func (p *Platform) CL() *cl.Platform {
	return p.ptr
}
