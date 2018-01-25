package worker

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
	"unsafe"

	"github.com/jgillich/go-opencl/cl"
	"github.com/pkg/errors"

	"github.com/GeertJohan/go.rice"
)

var cryptonightKernel string

func init() {
	box := rice.MustFindBox("../opencl")
	var out bytes.Buffer
	var re = regexp.MustCompile(`(#include "(.*\.cl)")`)
	cl := box.MustString("cryptonight.cl")
	cl = re.ReplaceAllString(cl, `{{ .MustString "$2" }}`)

	if err := template.Must(template.New("cryptonight").Parse(cl)).Execute(&out, box); err != nil {
		panic(err)
	}
	cryptonightKernel = out.String()
}

type CryptonightCLWorker struct {
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
	kernels       []*cl.Kernel
}

func NewCryptonightCLWorker(device *cl.Device, light bool) (*CryptonightCLWorker, error) {
	ctx, err := cl.CreateContext([]*cl.Device{device})
	if err != nil {
		return nil, err
	}

	var iterations int
	if light {
		iterations = CryptonightLiteIter
	} else {
		iterations = CryptonightIter
	}

	var mask int
	if light {
		mask = CryptonightLiteMask
	} else {
		mask = CryptonightMask
	}

	hashMemSize := CryptonightMemory
	computeUnits := device.MaxComputeUnits()

	// 224byte extra memory is used per thread for meta data
	maxIntensity := int(device.GlobalMemSize())/hashMemSize + 224

	// map intensity to a multiple of the compute unit count, 8 is the number of threads per work group
	intensity := (maxIntensity / (8 * computeUnits)) * computeUnits * 8

	// leave some free memory
	intensity--

	// TODO figure out the best maximum
	if intensity > 1000 {
		intensity = 1000
	}

	w := CryptonightCLWorker{
		intensity: int(intensity),
		nonce:     0,
		worksize:  device.MaxWorkGroupSize() / 8,
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

	w.scratchpadBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, CryptonightMemory*intensity)
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.hashStateBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 200*intensity)
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.blakeBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.blakeBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.groestlBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.jhBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.skeinBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*(intensity+2))
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	w.outputBuf, err = ctx.CreateEmptyBuffer(cl.MemReadWrite, 4*0x100)
	if err != nil {
		return nil, errors.Wrap(err, "intializing buffer failed")
	}

	program, err := ctx.CreateProgramWithSource([]string{cryptonightKernel})
	if err != nil {
		return nil, errors.Wrap(err, "creating program failed")
	}

	options := fmt.Sprintf("-DITERATIONS=%v -DMASK=%v -DWORKSIZE=%v -DSTRIDED_INDEX=%v", iterations, mask, w.worksize, 0)
	if err := program.BuildProgram([]*cl.Device{device}, options); err != nil {
		return nil, errors.Wrap(err, "building program failed")
	}

	w.kernels = []*cl.Kernel{}
	for _, name := range []string{"cn0", "cn1", "cn2", "Blake", "Groestl", "JH", "Skein"} {
		kernel, err := program.CreateKernel(name)
		if err != nil {
			return nil, errors.Wrap(err, "creating kernel failed")
		}

		w.kernels = append(w.kernels, kernel)
	}

	return &w, nil
}

func (w *CryptonightCLWorker) SetJob(input []byte, target uint64) error {

	// TODO ???
	// input[input_len] = 0x01;
	// memset(input + input_len + 1, 0, 88 - input_len - 1);

	uintensity := uint64(w.intensity)

	if _, err := w.queue.EnqueueWriteBuffer(w.inputBuf, true, 0, 88, unsafe.Pointer(&input[0]), nil); err != nil {
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

func (w *CryptonightCLWorker) RunJob() ([]byte, error) {

	// round up to next multiple of worksize
	threads := ((w.intensity + w.worksize - 1) / w.worksize) * w.worksize

	if threads%w.worksize != 0 {
		return nil, errors.New("threads is no multiple of workSize")
	}

	// TODO ???
	// size_t BranchNonces[4];
	// memset(BranchNonces,0,sizeof(size_t)*4);
	branchNonces := [4]byte{0, 0, 0, 0}

	zero := uint(0)

	// zero branch buffer counters
	{

		if _, err := w.queue.EnqueueWriteBuffer(w.blakeBuf, false, 4*w.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}

		if _, err := w.queue.EnqueueWriteBuffer(w.groestlBuf, false, 4*w.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}

		if _, err := w.queue.EnqueueWriteBuffer(w.jhBuf, false, 4*w.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}

		if _, err := w.queue.EnqueueWriteBuffer(w.skeinBuf, false, 4*w.intensity, 4, unsafe.Pointer(&zero), nil); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if _, err := w.queue.EnqueueWriteBuffer(w.outputBuf, false, 4*0xFF, 4, unsafe.Pointer(&zero), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := w.queue.Finish(); err != nil {
		return nil, errors.WithStack(err)
	}

	globalWorkOffset := []int{w.nonce, 1}
	globalWorkSize := []int{threads, 8}
	localWorkSize := []int{w.worksize, 8}

	if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[0], globalWorkOffset, globalWorkSize, localWorkSize, nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[1], []int{w.nonce}, []int{threads}, []int{w.worksize}, nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[2], globalWorkOffset, globalWorkSize, localWorkSize, nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.blakeBuf, false, 4*w.intensity, 4, unsafe.Pointer(&branchNonces[0]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.groestlBuf, false, 4*w.intensity, 4, unsafe.Pointer(&branchNonces[1]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.jhBuf, false, 4*w.intensity, 4, unsafe.Pointer(&branchNonces[2]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := w.queue.EnqueueReadBuffer(w.skeinBuf, false, 4*w.intensity, 4, unsafe.Pointer(&branchNonces[3]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := w.queue.Finish(); err != nil {
		return nil, errors.WithStack(err)
	}

	for i := 3; i < 7; i++ {
		ni := i - 3

		if branchNonces[ni] == 0 {
			continue
		}

		if err := w.kernels[i].SetArg(4, uint64(branchNonces[ni])); err != nil {
			return nil, errors.WithStack(err)
		}

		// round up to next multiple of workSize
		floatWorkSize := float32(w.worksize)
		branchNonces[ni] = byte(((float32(branchNonces[ni]) + floatWorkSize - 1) / floatWorkSize) * floatWorkSize)

		if int(branchNonces[ni])%w.worksize != 0 {
			return nil, errors.New("branchNonce is no multiple of workSize")
		}

		if _, err := w.queue.EnqueueNDRangeKernel(w.kernels[i], []int{w.nonce}, []int{int(branchNonces[ni])}, []int{w.worksize}, nil); err != nil {
			return nil, errors.WithStack(err)
		}

	}

	output := make([]byte, 4)
	if _, err := w.queue.EnqueueReadBuffer(w.outputBuf, true, 0, 4*0x100, unsafe.Pointer(&output[0]), nil); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := w.queue.Finish(); err != nil {
		return nil, errors.WithStack(err)
	}

	w.nonce += w.intensity

	return output, nil
}
