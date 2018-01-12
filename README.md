
coin is a easy to use miner for crypto currencies. It features automatic hardware detection,
support for many different algorithms and a optional graphical user interface.

Current state: Under development.

## Usage

Run `coin --help` to display usage. For command line usage, you first want to run `coin miner -init`
to generate the configuration file, and then use `coin miner` afterwards to start mining. To
launch the GUI, run `coin gui`.

## Building

Building coin is a multi step process and requires a lot of dependencies. At the moment, there
are no build instructions, please refer to `.gitlab-ci.yml` and `.appveyor.yml`.
