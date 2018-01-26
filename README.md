# BlockForge


[![GoDoc](https://godoc.org/gitlab.com/blockforge/blockforge?status.svg)](https://godoc.org/gitlab.com/blockforge/blockforge)
[![pipeline status](https://gitlab.com/blockforge/blockforge/badges/master/pipeline.svg)](https://gitlab.com/blockforge/blockforge/commits/master)
[![Build status](https://ci.appveyor.com/api/projects/status/6bl4w08cpa6163kx?svg=true)](https://ci.appveyor.com/project/JakobGillich/blockforge)

BlockForge is a easy to use miner for crypto currencies. It features automatic hardware detection,
support for many different algorithms and a optional graphical user interface.

Current state: Under development.

## Usage

Run `blockforge --help` to display usage. For command line usage, you first want to run `blockforge miner -init`
to generate the configuration file, and then use `blockforge miner` afterwards to start mining. To
launch the GUI, run `blockforge gui`.

## Building


### Linux

Install the following dependencies: GCC, cmake, Go, hwloc, webkit2gtk, ocl-icd

To build, run:

```
cmake -G -Hhash -Bhash/build
cmake --build hash/build
go get -u github.com/golang/dep/cmd/dep github.com/GeertJohan/go.rice/rice
dep ensure -vendor-only
go generate
go build
```

### Windows

Install the following dependencies: MSYS2, Go

Use the MSYS2 shell to install the rest of the dependencies:

```
pacman -S mingw-w64-x86_64-cmake mingw-w64-x86_64-gcc mingw-w64-x86_64-opencl-headers mingw-w64-x86_64-headers-git mingw-w64-x86_64-make
```

Download [hwloc](https://www.open-mpi.org/software/hwloc/v1.11/) and extract it to `C:\msys64\mingw64`.
Make sure `C:\msys64\mingw64\bin` is in your PATH.

To build, run:

```
cmake -G "MinGW Makefiles" -DCMAKE_PREFIX_PATH=c:\msys64\mingw64 -Hhash -Bhash/build
cmake --build hash/build
go get -u github.com/golang/dep/cmd/dep github.com/GeertJohan/go.rice/rice
dep ensure -vendor-only
go generate
go build
```
