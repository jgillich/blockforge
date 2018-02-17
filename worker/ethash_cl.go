package worker

import (
	"fmt"
	"math"
	"strings"

	"github.com/jgillich/go-opencl/cl"
	"github.com/pkg/errors"
	"gitlab.com/blockforge/blockforge/algo/ethash"

	"github.com/gobuffalo/packr"
)

type ethashCL struct {
	ctx          *cl.Context
	queue        *cl.CommandQueue
	program      *cl.Program
	cache        *cl.MemObject
	dag          *cl.MemObject
	header       *cl.MemObject
	search       *cl.MemObject
	searchKernel *cl.Kernel
	dagKernel    *cl.Kernel
}

func newEthashCL(config CLDeviceConfig, ethash *ethash.Ethash) (*ethashCL, error) {
	kernel, err := packr.NewBox("../opencl").MustString("ethash.cl")
	if err != nil {
		return nil, err
	}

	device := config.Device.CL()

	if int(config.Device.CL().GlobalMemSize()) < len(ethash.DAG) {
		return nil, fmt.Errorf("GPU has insufficient memory to fit DAG")
	}

	defaultLocalWorkSize := 128
	defaultGlobalWorkSizeMultiplier := 8192

	localWorkSize := ((defaultLocalWorkSize + 7) / 8) * 8
	workgroupSize := localWorkSize
	globalWorkSize := defaultGlobalWorkSizeMultiplier * localWorkSize
	globalWorkSize = ((globalWorkSize / workgroupSize) + 1) * workgroupSize

	ctx, err := cl.CreateContext([]*cl.Device{device})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queue, err := ctx.CreateCommandQueue(device, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	program, err := ctx.CreateProgramWithSource([]string{kernel})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	options := []string{
		fmt.Sprintf("-D%v=%v", "GROUP_SIZE", workgroupSize),
		fmt.Sprintf("-D%v=%v", "DAG_SIZE", len(ethash.DAG)/128),
		fmt.Sprintf("-D%v=%v", "LIGHT_SIZE", len(ethash.Cache)/64), // TODO what's the right size?
		//fmt.Sprintf("-D%v=%v", "ACCESSES", workgroupSize), TODO??
		fmt.Sprintf("-D%v=%v", "MAX_OUTPUTS", 1),
		// fmt.Sprintf("-D%v=%v", "PLATFORM", workgroupSize), TODO!!
		fmt.Sprintf("-D%v=%v", "COMPUTE", 0), // TODO
		fmt.Sprintf("-D%v=%v", "THREADS_PER_HASH", 8),
	}

	if err := program.BuildProgram([]*cl.Device{device}, strings.Join(options, " ")); err != nil {
		return nil, errors.WithStack(err)
	}

	cache, err := ctx.CreateBuffer(cl.MemReadOnly, ethash.Cache)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dag, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, len(ethash.DAG))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	searchKernel, err := program.CreateKernel("ethash_search")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dagKernel, err := program.CreateKernel("ethash_calculate_dag_item")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	header, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := searchKernel.SetArgBuffer(1, header); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := searchKernel.SetArgBuffer(2, dag); err != nil {
		return nil, errors.WithStack(err)
	}

	// from ethminer: "Pass this to stop the compiler unrolling the loops"
	if err := searchKernel.SetArgUint32(5, math.MaxUint32); err != nil {
		return nil, errors.WithStack(err)
	}

	search, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, (1+1)*4)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	work := len(ethash.DAG) / 128
	fullRuns := work / globalWorkSize
	restWork := work % globalWorkSize
	if restWork > 0 {
		fullRuns++
	}

	if err := dagKernel.SetArgBuffer(1, cache); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := dagKernel.SetArgBuffer(2, dag); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := dagKernel.SetArgUint32(3, math.MaxUint32); err != nil {
		return nil, errors.WithStack(err)
	}

	for i := 0; i < fullRuns; i++ {
		dagKernel.SetArg(0, i*globalWorkSize)
		if _, err := queue.EnqueueNDRangeKernel(dagKernel, nil, []int{globalWorkSize}, []int{workgroupSize}, nil); err != nil {
			return nil, errors.WithStack(err)
		}
		if err := queue.Finish(); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return &ethashCL{ctx, queue, program, cache, dag, header, search, searchKernel, dagKernel}, nil
}
