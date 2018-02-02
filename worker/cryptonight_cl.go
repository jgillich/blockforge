package worker

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
	"unsafe"

	"github.com/gobuffalo/packr"
	"github.com/jgillich/go-opencl/cl"
	"github.com/pkg/errors"
)

type cryptonightCLWorker struct {
	Intensity     uint32
	worksize      uint32
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

func newCryptonightCLWorker(config CLDeviceConfig, lite bool) (*cryptonightCLWorker, error) {

	var cryptonightKernel string
	{
		box := packr.NewBox("../opencl")

		var out bytes.Buffer
		cl, err := box.MustString("cryptonight.cl")
		if err != nil {
			return nil, errors.New("this binary is incomplete - did you run go generate?")
		}
		cl = regexp.MustCompile(`(#include "(.*\.cl)")`).ReplaceAllString(cl, `{{ .String "$2" }}`)
		if err := template.Must(template.New("cryptonight").Parse(cl)).Execute(&out, box); err != nil {
			panic(err)
		}
		cryptonightKernel = out.String()
	}

	device := config.Device.CL()

	ctx, err := cl.CreateContext([]*cl.Device{device})
	if err != nil {
		return nil, err
	}

	var memory uint32
	if lite {
		memory = CryptonightLiteMemory
	} else {
		memory = CryptonightMemory
	}

	var iterations uint32
	if lite {
		iterations = CryptonightLiteIter
	} else {
		iterations = CryptonightIter
	}

	var mask uint32
	if lite {
		mask = CryptonightLiteMask
	} else {
		mask = CryptonightMask
	}

	w := cryptonightCLWorker{
		Intensity: uint32(config.Intensity),
		worksize:  uint32(device.MaxWorkGroupSize() / 8),
	}

	if w.worksize == 0 {
		return nil, errors.New("workSize is 0")
	}

	w.queue, err = ctx.CreateCommandQueue(device, 0)
	if err != nil {
		return nil, errors.Wrap(err, "creating command queue failed")
	}

	w.inputBuf, err = ctx.CreateEmptyBuffer(cl.MemReadOnly, 88)
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.scratchpadBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, int(memory*w.Intensity))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.hashStateBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, int(200*w.Intensity))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.blakeBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, int(4*(w.Intensity+2)))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.groestlBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, int(4*(w.Intensity+2)))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.jhBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, int(4*(w.Intensity+2)))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.skeinBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, int(4*(w.Intensity+2)))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.outputBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*0x100)
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.program, err = ctx.CreateProgramWithSource([]string{cryptonightKernel})
	if err != nil {
		return nil, errors.Wrap(err, "creating program failed")
	}

	options := fmt.Sprintf("-DITERATIONS=%v -DMASK=%v -DWORKSIZE=%v -DSTRIDED_INDEX=%v", iterations, mask, w.worksize, 0)
	if err := w.program.BuildProgram([]*cl.Device{device}, options); err != nil {
		return nil, errors.Wrap(err, "building program failed")
	}

	w.kernels = []*cl.Kernel{}
	for _, name := range []string{"cn0", "cn1", "cn2", "Blake", "Groestl", "JH", "Skein"} {
		kernel, err := w.program.CreateKernel(name)
		if err != nil {
			return nil, errors.Wrap(err, "creating kernel failed")
		}

		w.kernels = append(w.kernels, kernel)
	}

	return &w, nil
}

func (w *cryptonightCLWorker) SetJob(input []byte, target uint64) error {

	uintensity := uint64(w.Intensity)

	in := make([]byte, 88)
	for i := 0; i < len(input); i++ {
		in[i] = input[i]
	}
	in[len(input)] = 0x01

	if _, err := w.queue.EnqueueWriteBuffer(w.inputBuf, true, 0, 88, unsafe.Pointer(&in[0]), nil); err != nil {
		return errors.WithStack(err)
	}

	// kernel cn0

	if err := w.kernels[0].SetArgBuffer(0, w.inputBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[0].SetArgBuffer(1, w.scratchpadBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[0].SetArgBuffer(2, w.hashStateBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[0].SetArg(3, uintensity); err != nil {
		return errors.WithStack(err)
	}

	// kernel cn1

	if err := w.kernels[1].SetArgBuffer(0, w.scratchpadBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[1].SetArgBuffer(1, w.hashStateBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[1].SetArg(2, uintensity); err != nil {
		return errors.WithStack(err)
	}

	// kernel cn2

	if err := w.kernels[2].SetArgBuffer(0, w.scratchpadBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[2].SetArgBuffer(1, w.hashStateBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[2].SetArgBuffer(2, w.blakeBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[2].SetArgBuffer(3, w.groestlBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[2].SetArgBuffer(4, w.jhBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[2].SetArgBuffer(5, w.skeinBuf); err != nil {
		return errors.WithStack(err)
	}

	if err := w.kernels[2].SetArg(6, uintensity); err != nil {
		return errors.WithStack(err)
	}

	for i := 3; i < 7; i++ {
		if err := w.kernels[i].SetArgBuffer(0, w.hashStateBuf); err != nil {
			return errors.WithStack(err)
		}

		switch i {
		case 3:
			if err := w.kernels[i].SetArgBuffer(1, w.blakeBuf); err != nil {
				return errors.WithStack(err)
			}
		case 4:
			if err := w.kernels[i].SetArgBuffer(1, w.groestlBuf); err != nil {
				return errors.WithStack(err)
			}
		case 5:
			if err := w.kernels[i].SetArgBuffer(1, w.jhBuf); err != nil {
				return errors.WithStack(err)
			}
		case 6:
			if err := w.kernels[i].SetArgBuffer(1, w.skeinBuf); err != nil {
				return errors.WithStack(err)
			}
		}

		if err := w.kernels[i].SetArgBuffer(2, w.outputBuf); err != nil {
			return errors.WithStack(err)
		}

		if err := w.kernels[i].SetArg(3, target); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (w *cryptonightCLWorker) RunJob(results []uint32, nonce uint32) error {

	// round up to next multiple of worksize
	threads := ((w.Intensity + w.worksize - 1) / w.worksize) * w.worksize

	if threads%w.worksize != 0 {
		return errors.New("threads is no multiple of workSize")
	}

	branchNonces := make([]uint32, 4)

	zero := uint32(0)

	// zero branch buffer counters
	{

		if _, err := w.queue.EnqueueWriteBuffer(w.blakeBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&zero), nil); err != nil {
			return errors.WithStack(err)
		}

		if _, err := w.queue.EnqueueWriteBuffer(w.groestlBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&zero), nil); err != nil {
			return errors.WithStack(err)
		}

		if _, err := w.queue.EnqueueWriteBuffer(w.jhBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&zero), nil); err != nil {
			return errors.WithStack(err)
		}

		if _, err := w.queue.EnqueueWriteBuffer(w.skeinBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&zero), nil); err != nil {
			return errors.WithStack(err)
		}
	}

	if _, err := w.queue.EnqueueWriteBuffer(w.outputBuf, false, 4*0xFF, 4, unsafe.Pointer(&zero), nil); err != nil {
		return errors.WithStack(err)
	}

	if err := w.queue.Finish(); err != nil {
		return errors.WithStack(err)
	}

	globalWorkOffset := []int{int(nonce), 1}
	globalWorkSize := []int{int(threads), 8}
	localWorkSize := []int{int(w.worksize), 8}

	if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[0], globalWorkOffset, globalWorkSize, localWorkSize, nil); err != nil {
		return errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[1], []int{int(nonce)}, []int{int(threads)}, []int{int(w.worksize)}, nil); err != nil {
		return errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[2], globalWorkOffset, globalWorkSize, localWorkSize, nil); err != nil {
		return errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.blakeBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&branchNonces[0]), nil); err != nil {
		return errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.groestlBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&branchNonces[1]), nil); err != nil {
		return errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.jhBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&branchNonces[2]), nil); err != nil {
		return errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.skeinBuf, false, int(4*w.Intensity), 4, unsafe.Pointer(&branchNonces[3]), nil); err != nil {
		return errors.WithStack(err)
	}

	if err := w.queue.Finish(); err != nil {
		return errors.WithStack(err)
	}

	for i := 3; i < 7; i++ {
		ni := i - 3

		if branchNonces[ni] == 0 {
			continue
		}

		if err := w.kernels[i].SetArg(4, uint64(branchNonces[ni])); err != nil {
			return errors.WithStack(err)
		}

		// round up to next multiple of workSize
		branchNonces[ni] = ((branchNonces[ni] + w.worksize - 1) / w.worksize) * w.worksize

		if branchNonces[ni]%w.worksize != 0 {
			return errors.New("branchNonce is no multiple of workSize")
		}

		if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[i], []int{int(nonce)}, []int{int(branchNonces[ni])}, []int{int(w.worksize)}, nil); err != nil {
			return errors.WithStack(err)
		}

	}

	if _, err := w.queue.EnqueueReadBuffer(w.outputBuf, true, 0, 4*0x100, unsafe.Pointer(&results[0]), nil); err != nil {
		return errors.WithStack(err)
	}

	if err := w.queue.Finish(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (w *cryptonightCLWorker) Release() {
	w.queue.Release()
	w.inputBuf.Release()
	w.scratchpadBuf.Release()
	w.hashStateBuf.Release()
	w.blakeBuf.Release()
	w.groestlBuf.Release()
	w.jhBuf.Release()
	w.outputBuf.Release()
	for _, k := range w.kernels {
		k.Release()
	}
	w.program.Release()
}
