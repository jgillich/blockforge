## Features

coin is a easy to use miner for cryptocurrencies.

Currently supported are:

* Ethereum (GPU only)
* Monero

## Usage

Run `coin miner --init` to generate the initial config file, then modify it to your liking.
At the very least, change the pool configuration and make sure all CPUs and GPUs are configured
properly.

## Building

First, install the following dependencies:

* Go & dep
* cmake and a C++ compiler
* OpenSSL
* Boost
* hwloc
* swig

Next, build all miners:

    cd cgo/xmrstak/xmr-stak
    mkdir build && cd build
    cmake .. -DCUDA_ENABLE=OFF -DMICROHTTPD_ENABLE=OFF
    cmake --build .

    cd cgo/ethminer/ethminer
    mkdir build && cd build
    cmake ..
    cmake --build .

To install Go dependencies, run `dep ensure`. To build, run `go build`.
