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

Dependencies:

* Go & dep
* OpenSSL
* Boost
* hwloc

To install dependencies, run `dep ensure`. To build, run `go build`.