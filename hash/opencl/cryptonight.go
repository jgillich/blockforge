package opencl

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/jgillich/go-opencl/cl"

	"github.com/GeertJohan/go.rice"
)

var CryptonightMemory = 2097152
var CryptonightMask = 0x1FFFF0
var CryptonightIter = 0x80000

var CryptonightLiteMemory = 1048576
var CryptonightLiteMask = 0xFFFF0
var CryptonightLiteIter = 0x40000

var cryptonightKernel string

func init() {
	var box = rice.MustFindBox(".")
	var out bytes.Buffer

	err := template.Must(template.New("cryptonight").Parse(box.MustString("cryptonight.cl"))).Execute(&out, box)
	if err != nil {
		panic(err)
	}

	cryptonightKernel = out.String()
}

type Cryptonight struct {
	ctx  *cl.Context
	gpus []CryptonightGpuContext
}

type CryptonightGpuContext struct {
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

func NewCryptonight(devices []*cl.Device) (*Cryptonight, error) {
	intensity := 1
	ctx, err := cl.CreateContext(devices)
	if err != nil {
		return nil, err
	}

	cryptonight := Cryptonight{
		ctx:  ctx,
		gpus: []CryptonightGpuContext{},
	}

	for _, device := range devices {
		gpuCtx := CryptonightGpuContext{}

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
