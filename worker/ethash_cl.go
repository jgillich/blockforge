package worker

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"unsafe"

	"github.com/jgillich/go-opencl/cl"
	"github.com/pkg/errors"
	"gitlab.com/blockforge/blockforge/algo/ethash"

	"github.com/gobuffalo/packr"
)

type ethashCL struct {
	ctx            *cl.Context
	queue          *cl.CommandQueue
	program        *cl.Program
	cache          *cl.MemObject
	dag            *cl.MemObject
	header         *cl.MemObject
	search         *cl.MemObject
	searchKernel   *cl.Kernel
	dagKernel      *cl.Kernel
	localWorkSize  int
	workgroupSize  int
	globalWorkSize int
}

func newEthashCL(config CLDeviceConfig, light *ethash.Light) (*ethashCL, error) {
	kernel, err := packr.NewBox("../opencl").MustString("ethash.cl")
	if err != nil {
		return nil, err
	}

	device := config.Device.CL()

	if int(config.Device.CL().GlobalMemSize()) < light.DataSize {
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

	// TODO CreateBuffer results in Invalid Host Ptr, might be a bug in the bindings
	cache, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, len(light.Cache))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if _, err := queue.EnqueueWriteBuffer(cache, true, 0, len(light.Cache), unsafe.Pointer(&light.Cache[0]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	dag, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, light.DataSize/4)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	header, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	program, err := ctx.CreateProgramWithSource([]string{kernel})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	options := []string{
		fmt.Sprintf("-D%v=%v", "PLATFORM", 0), // TODO 1 for AMD, 2 for NVIDIA
		fmt.Sprintf("-D%v=%v", "GROUP_SIZE", workgroupSize),
		fmt.Sprintf("-D%v=%v", "DAG_SIZE", light.DataSize/128/4),
		fmt.Sprintf("-D%v=%v", "LIGHT_SIZE", len(light.Cache)/64), // TODO what's the right size?
		//fmt.Sprintf("-D%v=%v", "ACCESSES", workgroupSize), TODO??
		fmt.Sprintf("-D%v=%v", "MAX_OUTPUTS", "1u"),
		// fmt.Sprintf("-D%v=%v", "PLATFORM", workgroupSize), TODO!!
		fmt.Sprintf("-D%v=%v", "COMPUTE", 0), // TODO
		fmt.Sprintf("-D%v=%v", "THREADS_PER_HASH", 8),
	}

	if err := program.BuildProgram([]*cl.Device{device}, strings.Join(options, " ")); err != nil {
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

	work := light.DataSize / 128 / 4
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
		dagKernel.SetArgUint32(0, uint32(i*globalWorkSize))
		if _, err := queue.EnqueueNDRangeKernel(dagKernel, nil, []int{globalWorkSize}, []int{workgroupSize}, nil); err != nil {
			return nil, errors.WithStack(err)
		}
		if err := queue.Finish(); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return &ethashCL{
		ctx:            ctx,
		queue:          queue,
		program:        program,
		cache:          cache,
		dag:            dag,
		header:         header,
		search:         search,
		searchKernel:   searchKernel,
		dagKernel:      dagKernel,
		localWorkSize:  localWorkSize,
		workgroupSize:  workgroupSize,
		globalWorkSize: globalWorkSize,
	}, nil
}

func (cl *ethashCL) Update(header []byte, target *big.Int) error {
	zero := uint32(0)
	targetBytes := target.Bytes()

	if _, err := cl.queue.EnqueueWriteBuffer(cl.header, false, 0, len(header), unsafe.Pointer(&header[0]), nil); err != nil {
		return errors.WithStack(err)
	}

	if _, err := cl.queue.EnqueueWriteBuffer(cl.search, false, 0, 4, unsafe.Pointer(&zero), nil); err != nil {
		return errors.WithStack(err)
	}

	if err := cl.searchKernel.SetArgBuffer(0, cl.search); err != nil {
		return errors.WithStack(err)
	}

	if err := cl.searchKernel.SetArgUnsafe(4, 8, unsafe.Pointer(&targetBytes[0])); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (cl *ethashCL) Run(nonce uint64, results [2]uint32) error {

	if _, err := cl.queue.EnqueueReadBuffer(cl.search, true, 0, 4*len(results), unsafe.Pointer(&results[0]), nil); err != nil {
		return errors.WithStack(err)
	}

	if err := cl.searchKernel.SetArgUint64(3, nonce); err != nil {
		return errors.WithStack(err)
	}

	if _, err := cl.queue.EnqueueNDRangeKernel(cl.searchKernel, nil, []int{cl.globalWorkSize}, []int{cl.workgroupSize}, nil); err != nil {
		return errors.WithStack(err)
	}

	if err := cl.queue.Finish(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (cl *ethashCL) Release() {
	defer cl.ctx.Release()
	defer cl.queue.Release()
	defer cl.program.Release()
	defer cl.cache.Release()
	defer cl.dag.Release()
	defer cl.header.Release()
	defer cl.search.Release()
	defer cl.searchKernel.Release()
	defer cl.dagKernel.Release()
}
