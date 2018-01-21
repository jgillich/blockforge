
coin is a easy to use miner for crypto currencies. It features automatic hardware detection,
support for many different algorithms and a optional graphical user interface.

Current state: Under development.

## Usage

Run `coin --help` to display usage. For command line usage, you first want to run `coin miner -init`
to generate the configuration file, and then use `coin miner` afterwards to start mining. To
launch the GUI, run `coin gui`.

## Building


### Linux

Install the following dependencies: GCC, cmake, Go, hwloc, webkit2gtk, ocl-icd

To build, run:

```
cmake -G -Hhash -Bhash/build
cmake --build hash/build
go get -u github.com/golang/dep/cmd/dep
go get -u github.com/GeertJohan/go.rice/rice
dep ensure -vendor-only
rice embed-go -i ./cmd
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
go get -u github.com/golang/dep/cmd/dep
go get -u github.com/GeertJohan/go.rice/rice
dep ensure -vendor-only
rice embed-go -i ./cmd
go build
```
