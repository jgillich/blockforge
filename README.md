# BlockForge

[![GoDoc](https://godoc.org/gitlab.com/blockforge/blockforge?status.svg)](https://godoc.org/gitlab.com/blockforge/blockforge)
[![pipeline status](https://gitlab.com/blockforge/blockforge/badges/master/pipeline.svg)](https://gitlab.com/blockforge/blockforge/commits/master)
[![Build status](https://ci.appveyor.com/api/projects/status/6bl4w08cpa6163kx?svg=true)](https://ci.appveyor.com/project/JakobGillich/blockforge)

 BlockForge is a next generation miner for cryptocurrencies.
 Easy to use, multi algo and open source.

Current state: Under development.

## Usage

Run `blockforge --help` to display usage.

## Building


### Linux

Install the following dependencies: GCC, cmake, Go, hwloc, webkit2gtk, ocl-icd

To build, run:

```
cmake -Hhash -Bhash/build
cmake --build hash/build
go get -u github.com/golang/dep/cmd/dep github.com/gobuffalo/packr/...
dep ensure -vendor-only
go generate
go build
```

### Windows

Install the following dependencies: MSYS2, Go, Git

Use the MSYS2 shell to install the rest of the dependencies:

```
pacman -S mingw-w64-x86_64-cmake mingw-w64-x86_64-gcc mingw-w64-x86_64-opencl-headers mingw-w64-x86_64-headers-git mingw-w64-x86_64-make
```

Download [hwloc](https://www.open-mpi.org/software/hwloc/v1.11/) and extract it to `C:\msys64\mingw64`.
Extract the x64 `OpenCL.lib` from [OCL-SDK](https://github.com/GPUOpen-LibrariesAndSDKs/OCL-SDK/releases) to `C:\msys64\mingw64\lib`.
Make sure `C:\msys64\mingw64\bin` is in your PATH.

To build, run:

```
cmake -G "MinGW Makefiles" -DCMAKE_PREFIX_PATH=c:\msys64\mingw64 -Hhash -Bhash/build
cmake --build hash/build
go get -u github.com/golang/dep/cmd/dep github.com/gobuffalo/packr/...
dep ensure -vendor-only
go generate
go build -ldflags '-linkmode external -extldflags "-static"'
```
