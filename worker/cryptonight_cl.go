package worker

import (
	"bytes"
	"fmt"
	"text/template"
	"unsafe"

	"github.com/jgillich/go-opencl/cl"

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
	threads       int
	workSize      int
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
			threads: intensity,
			nonce:   0,
			// TODO what to set this to?
			workSize: device.MaxWorkGroupSize() / 8,
		}

		if gpuCtx.workSize == 0 {
			panic("workSize is 0")
		}

		gpuCtx.queue, err = ctx.CreateCommandQueue(device, 0)
		if err != nil {
			return nil, err
		}

		gpuCtx.inputBuf, err = ctx.CreateEmptyBuffer(cl.MemReadOnly, 88)
		if err != nil {
			return nil, err
		}

		gpuCtx.scratchpadBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, CryptonightMemory*intensity)
		if err != nil {
			return nil, err
		}

		gpuCtx.hashStateBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 200*intensity)
		if err != nil {
			return nil, err
		}

		gpuCtx.blakeBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, err
		}

		gpuCtx.blakeBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, err
		}

		gpuCtx.groestlBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, err
		}

		gpuCtx.jhBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, err
		}

		gpuCtx.skeinBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
		if err != nil {
			return nil, err
		}

		gpuCtx.outputBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*0x100)
		if err != nil {
			return nil, err
		}

		gpuCtx.program, err = ctx.CreateProgramWithSource([]string{cryptonightKernel})
		if err != nil {
			return nil, err
		}

		options := fmt.Sprintf("-DITERATIONS=%v -DMASK=%v -DWORKSIZE=%v -DSTRIDED_INDEX=%v", CryptonightIter, CryptonightMask, CryptonightMemory, 0)
		err = gpuCtx.program.BuildProgram([]*cl.Device{device}, options)
		if err != nil {
			return nil, err
		}
		gpuCtx.kernels = []*cl.Kernel{}

		for _, name := range []string{"cn0", "cn1", "cn2", "Blake", "Groestl", "JH", "Skein"} {
			kernel, err := gpuCtx.program.CreateKernel(name)
			if err != nil {
				return nil, err
			}

			gpuCtx.kernels = append(gpuCtx.kernels, kernel)
		}

		cryptonight.gpus = append(cryptonight.gpus, gpuCtx)
	}

	return &cryptonight, nil
}

func (ctx *CryptonightCLContext) SetJob(input []byte, target uint) error {

	// TODO ???
	// input[input_len] = 0x01;
	// memset(input + input_len + 1, 0, 88 - input_len - 1);

	_, err := ctx.queue.EnqueueWriteBuffer(ctx.inputBuf, true, 0, 88, unsafe.Pointer(&input[0]), nil)
	if err != nil {
		return err
	}

	// kernel cn0

	err = ctx.kernels[0].SetArgBuffer(0, ctx.inputBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[0].SetArgBuffer(1, ctx.scratchpadBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[0].SetArgBuffer(2, ctx.hashStateBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[0].SetArg(3, &ctx.threads)
	if err != nil {
		return err
	}

	// kernel cn1

	err = ctx.kernels[1].SetArgBuffer(0, ctx.scratchpadBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[1].SetArgBuffer(1, ctx.hashStateBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[1].SetArg(2, &ctx.threads)
	if err != nil {
		return err
	}

	// kernel cn2

	err = ctx.kernels[2].SetArgBuffer(0, ctx.scratchpadBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[2].SetArgBuffer(1, ctx.hashStateBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[2].SetArgBuffer(2, ctx.blakeBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[2].SetArgBuffer(3, ctx.groestlBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[2].SetArgBuffer(4, ctx.jhBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[2].SetArgBuffer(5, ctx.skeinBuf)
	if err != nil {
		return err
	}

	err = ctx.kernels[2].SetArg(6, &ctx.threads)
	if err != nil {
		return err
	}

	for i := 3; i < 7; i++ {
		err = ctx.kernels[i].SetArgBuffer(0, ctx.hashStateBuf)
		if err != nil {
			return err
		}

		switch i {
		case 3:
			err = ctx.kernels[i].SetArgBuffer(1, ctx.blakeBuf)
			if err != nil {
				return err
			}
		case 4:
			err = ctx.kernels[i].SetArgBuffer(1, ctx.groestlBuf)
			if err != nil {
				return err
			}
		case 5:
			err = ctx.kernels[i].SetArgBuffer(1, ctx.jhBuf)
			if err != nil {
				return err
			}
		case 6:
			err = ctx.kernels[i].SetArgBuffer(1, ctx.skeinBuf)
			if err != nil {
				return err
			}
		}

		err = ctx.kernels[i].SetArgBuffer(2, ctx.outputBuf)
		if err != nil {
			return err
		}

		err = ctx.kernels[i].SetArg(3, &target)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *CryptonightCLContext) RunJob() error {

	// round up to next multiple of w_size
	threads := ((ctx.threads + ctx.workSize - 1) / ctx.workSize) * ctx.workSize

	if threads%ctx.workSize == 0 {
		panic("threads is no multiple of workSize")
	}

	// TODO ???
	// size_t BranchNonces[4];
	// memset(BranchNonces,0,sizeof(size_t)*4);
	branchNonces := [4]byte{0}

	zero := 0

	// zero branch buffer counters
	{

		_, err := ctx.queue.EnqueueWriteBuffer(ctx.blakeBuf, false, 4*threads, 4, unsafe.Pointer(&zero), nil)
		if err != nil {
			return err
		}

		_, err = ctx.queue.EnqueueWriteBuffer(ctx.groestlBuf, false, 4*threads, 4, unsafe.Pointer(&zero), nil)
		if err != nil {
			return err
		}

		_, err = ctx.queue.EnqueueWriteBuffer(ctx.jhBuf, false, 4*threads, 4, unsafe.Pointer(&zero), nil)
		if err != nil {
			return err
		}

		_, err = ctx.queue.EnqueueWriteBuffer(ctx.skeinBuf, false, 4*threads, 4, unsafe.Pointer(&zero), nil)
		if err != nil {
			return err
		}
	}

	_, err := ctx.queue.EnqueueWriteBuffer(ctx.outputBuf, false, 4*0xFF, 4, unsafe.Pointer(&zero), nil)
	if err != nil {
		return err
	}

	err = ctx.queue.Finish()
	if err != nil {
		return err
	}

	nonce := []int{ctx.nonce, 1}
	thr := []int{threads, 8}
	wrk := []int{ctx.workSize, 8}

	_, err = ctx.queue.EnqueueNDRangeKernel(ctx.kernels[0], nonce, thr, wrk, nil)
	if err != nil {
		return err
	}

	_, err = ctx.queue.EnqueueNDRangeKernel(ctx.kernels[1], []int{ctx.nonce}, []int{threads}, []int{ctx.workSize}, nil)
	if err != nil {
		return err
	}

	_, err = ctx.queue.EnqueueNDRangeKernel(ctx.kernels[2], nonce, thr, wrk, nil)
	if err != nil {
		return err
	}

	_, err = ctx.queue.EnqueueReadBuffer(ctx.blakeBuf, false, 4*ctx.threads, 4, unsafe.Pointer(&branchNonces), nil)
	if err != nil {
		return err
	}

	_, err = ctx.queue.EnqueueReadBuffer(ctx.groestlBuf, false, 4*ctx.threads, 4, unsafe.Pointer(&branchNonces[1]), nil)
	if err != nil {
		return err
	}

	_, err = ctx.queue.EnqueueReadBuffer(ctx.jhBuf, false, 4*ctx.threads, 4, unsafe.Pointer(&branchNonces[2]), nil)
	if err != nil {
		return err
	}

	_, err = ctx.queue.EnqueueReadBuffer(ctx.skeinBuf, false, 4*ctx.threads, 4, unsafe.Pointer(&branchNonces[2]), nil)
	if err != nil {
		return err
	}

	err = ctx.queue.Finish()
	if err != nil {
		return err
	}

	for i := 3; i < 7; i++ {
		err = ctx.kernels[i].SetArg(4, &branchNonces[i])
		if err != nil {
			return err
		}

		// round up to next multiple of workSize
		branchNonces[i] = byte(((int(branchNonces[i]) + ctx.workSize - 1) / ctx.workSize) * ctx.workSize)

		if int(branchNonces[i])%ctx.workSize == 0 {
			panic("branchNonce is no multiple of workSize")
		}

		_, err = ctx.queue.EnqueueNDRangeKernel(ctx.kernels[i], []int{ctx.nonce}, []int{int(branchNonces[i])}, []int{ctx.workSize}, nil)
		if err != nil {
			return err
		}

	}

	return nil
}
