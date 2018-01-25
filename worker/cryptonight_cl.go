package worker

import (
	"bytes"
	"fmt"
	"text/template"
	"unsafe"

	"github.com/jgillich/go-opencl/cl"
	"github.com/pkg/errors"

	"github.com/GeertJohan/go.rice"
)

var cryptonightKernel string

func init() {
	var box = rice.MustFindBox("../opencl")
	var out bytes.Buffer

	err := template.Must(template.New("cryptonight").Parse(box.MustString("cryptonight.cl"))).Execute(&out, box)
	if err != nil {
		panic(err)
	}

	cryptonightKernel = out.String()
}

type CryptonightCLWorker struct {
	ctx  *cl.Context
	gpus []CryptonightCLContext
}

type CryptonightCLContext struct {
	intensity     int
	worksize      int
	nonce         int
	queue         *cl.CommandQueue
	inputBuf      *cl.MemObject
	scratchpadBuf *cl.MemObject
	hashStateBuf  *cl.MemObject
	blakeBuf      *cl.MemObject
	groestlBuf    *cl.MemObject
	jhBuf         *cl.MemObject
	skeinBuf      *cl.MemObject
	outputBuf     *cl.MemObject
	program       *cl.Program
	kernels       []*cl.Kernel
}

func NewCryptonightCLWorker(devices []*cl.Device) (*CryptonightCLWorker, error) {
	intensity := 1
	ctx, err := cl.CreateContext(devices)
	if err != nil {
		return nil, err
	}

	cryptonight := CryptonightCLWorker{
		ctx:  ctx,
		gpus: []CryptonightCLContext{},
	}

	for _, device := range devices {
		gpuCtx := CryptonightCLContext{
			intensity: intensity,
			nonce:     0,
			// TODO what to set this to?
			worksize: device.MaxWorkGroupSize() / 8,
		}

		if gpuCtx.worksize == 0 {
			panic("workSize is 0")
		}

		gpuCtx.queue, err = ctx.CreateCommandQueue(device, 0)
		if err != nil {
			return nil, errors.Wrap(err, "creating command queue failed")
		}

		gpuCtx.inputBuf, err = ctx.CreateEmptyBuffer(cl.MemReadOnly, 88)
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.scratchpadBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, CryptonightMemory*intensity)
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.hashStateBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 200*intensity)
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.blakeBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.blakeBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.groestlBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.jhBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.skeinBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.outputBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*0x100)
		if err != nil {
			return nil, errors.Wrap(err, "intializing buffer failed")
		}

		gpuCtx.program, err = ctx.CreateProgramWithSource([]string{cryptonightKernel})
		if err != nil {
			return nil, errors.Wrap(err, "creating program failed")
		}

		options := fmt.Sprintf("-DITERATIONS=%v -DMASK=%v -DWORKSIZE=%v -DSTRIDED_INDEX=%v", CryptonightIter, CryptonightMask, CryptonightMemory, 0)
		err = gpuCtx.program.BuildProgram([]*cl.Device{device}, options)
		if err != nil {
			return nil, errors.Wrap(err, "building program failed")
		}
		gpuCtx.kernels = []*cl.Kernel{}

		for _, name := range []string{"cn0", "cn1", "cn2", "Blake", "Groestl", "JH", "Skein"} {
			kernel, err := gpuCtx.program.CreateKernel(name)
			if err != nil {
				return nil, errors.Wrap(err, "creating kernel failed")
			}

			gpuCtx.kernels = append(gpuCtx.kernels, kernel)
		}

		cryptonight.gpus = append(cryptonight.gpus, gpuCtx)
	}

	return &cryptonight, nil
}

func (ctx *CryptonightCLContext) SetJob(input []byte, target uint64) error {

	// TODO ???
	// input[input_len] = 0x01;
	// memset(input + input_len + 1, 0, 88 - input_len - 1);

	uintensity := uint64(ctx.intensity)

	if _, err := ctx.queue.EnqueueWriteBuffer(ctx.inputBuf, true, 0, 88, unsafe.Pointer(&input[0]), nil); err != nil {
		return errors.WithStack(err)
	}

	// kernel cn0

	if err := ctx.kernels[0].SetArgBuffer(0, ctx.inputBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[0].SetArgBuffer(1, ctx.scratchpadBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[0].SetArgBuffer(2, ctx.hashStateBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[0].SetArg(3, uintensity); err != nil {
		return errors.WithStack(err)
	}

	// kernel cn1

	if err := ctx.kernels[1].SetArgBuffer(0, ctx.scratchpadBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[1].SetArgBuffer(1, ctx.hashStateBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[1].SetArg(2, uintensity); err != nil {
		return errors.WithStack(err)
	}

	// kernel cn2

	if err := ctx.kernels[2].SetArgBuffer(0, ctx.scratchpadBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[2].SetArgBuffer(1, ctx.hashStateBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[2].SetArgBuffer(2, ctx.blakeBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[2].SetArgBuffer(3, ctx.groestlBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[2].SetArgBuffer(4, ctx.jhBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[2].SetArgBuffer(5, ctx.skeinBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.kernels[2].SetArg(6, uintensity); err != nil {
		return errors.WithStack(err)
	}

	for i := 3; i < 7; i++ {
		if err := ctx.kernels[i].SetArgBuffer(0, ctx.hashStateBuf); err != nil {
			return errors.WithStack(err)
		}

		switch i {
		case 3:
			if err := ctx.kernels[i].SetArgBuffer(1, ctx.blakeBuf); err != nil {
				return errors.WithStack(err)
			}
		case 4:
			if err := ctx.kernels[i].SetArgBuffer(1, ctx.groestlBuf); err != nil {
				return errors.WithStack(err)
			}
		case 5:
			if err := ctx.kernels[i].SetArgBuffer(1, ctx.jhBuf); err != nil {
				return errors.WithStack(err)
			}
		case 6:
			if err := ctx.kernels[i].SetArgBuffer(1, ctx.skeinBuf); err != nil {
				return errors.WithStack(err)
			}
		}

		if err := ctx.kernels[i].SetArgBuffer(2, ctx.outputBuf); err != nil {
			return errors.WithStack(err)
		}

		if err := ctx.kernels[i].SetArg(3, target); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (ctx *CryptonightCLContext) RunJob() ([]byte, error) {

	// round up to next multiple of worksize
	threads := ((ctx.intensity + ctx.worksize - 1) / ctx.worksize) * ctx.worksize

	if threads%ctx.worksize != 0 {
		return nil, errors.WithStack(errors.New("threads is no multiple of workSize"))
	}

	// TODO ???
	// size_t BranchNonces[4];
	// memset(BranchNonces,0,sizeof(size_t)*4);
	branchNonces := [4]byte{0}

	zero := uint(0)

	// zero branch buffer counters
	{

		if _, err := ctx.queue.EnqueueWriteBuffer(ctx.blakeBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}

		if _, err := ctx.queue.EnqueueWriteBuffer(ctx.groestlBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}

		if _, err := ctx.queue.EnqueueWriteBuffer(ctx.jhBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}

		if _, err := ctx.queue.EnqueueWriteBuffer(ctx.skeinBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if _, err := ctx.queue.EnqueueWriteBuffer(ctx.outputBuf, false, 4*0xFF, 4, unsafe.Pointer(&zero), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := ctx.queue.Finish(); err != nil {
		return nil, errors.WithStack(err)
	}

	nonce := []int{ctx.nonce, 1}
	thr := []int{threads, 8}
	wrk := []int{ctx.worksize, 8}

	if _, err := ctx.queue.EnqueueNDRangeKernel(ctx.kernels[0], nonce, thr, wrk, nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := ctx.queue.EnqueueNDRangeKernel(ctx.kernels[1], []int{ctx.nonce}, []int{threads}, []int{ctx.worksize}, nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := ctx.queue.EnqueueNDRangeKernel(ctx.kernels[2], nonce, thr, wrk, nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := ctx.queue.EnqueueReadBuffer(ctx.blakeBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&branchNonces), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := ctx.queue.EnqueueReadBuffer(ctx.groestlBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&branchNonces[1]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := ctx.queue.EnqueueReadBuffer(ctx.jhBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&branchNonces[2]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := ctx.queue.EnqueueReadBuffer(ctx.skeinBuf, false, 4*ctx.intensity, 4, unsafe.Pointer(&branchNonces[2]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := ctx.queue.Finish(); err != nil {
		return nil, errors.WithStack(err)
	}

	for i := 3; i < 7; i++ {
		if err := ctx.kernels[i].SetArg(4, &branchNonces[i]); err != nil {
			return nil, errors.WithStack(err)
		}

		// round up to next multiple of workSize
		branchNonces[i] = byte(((int(branchNonces[i]) + ctx.worksize - 1) / ctx.worksize) * ctx.worksize)

		if int(branchNonces[i])%ctx.worksize != 0 {
			return nil, errors.WithStack(errors.New("branchNonce is no multiple of workSize"))
		}

		if _, err := ctx.queue.EnqueueNDRangeKernel(ctx.kernels[i], []int{ctx.nonce}, []int{int(branchNonces[i])}, []int{ctx.worksize}, nil); err != nil {
			return nil, errors.WithStack(err)
		}

	}

	output := make([]byte, 4)
	if _, err := ctx.queue.EnqueueReadBuffer(ctx.outputBuf, true, 0, 4*0x100, unsafe.Pointer(&output), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := ctx.queue.Finish(); err != nil {
		return nil, errors.WithStack(err)
	}

	ctx.nonce += threads

	return output, nil
}
