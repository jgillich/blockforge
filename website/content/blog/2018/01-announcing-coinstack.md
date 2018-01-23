---
title: "Announcing CoinStack, a modern multi miner"
date: 2018-01-21
---

Let's face it, the current state of software for cryptocurrency mining is less than ideal.
There are easy to use miners that support a wide range of coins but are closed source
and often accused of fraudulent behaviour. Then there are open source miners that tend to
perform much better, but are difficult to use and often only support a single algorithm or hardware
type.

CoinStack is an attempt to create a miner that ticks all the boxes:

* Easy to use
* Multiple algorithms
* CPU and GPU mining
* Open source

Our inital plan was to integrate existing open source miners, but for technical reasons we
had to

CoinStack is a modern miner featuring automatic hardware detection, support for many different
coins and a optional graphical user interface. And it is of course open source!

Our original plan was to provide a simpler interface
around existing open source miners, but for various
reasons that wasn't practical. Instead, we've implemented
a Stratum miner from scratch. Because we wanted
to involve the community as early as possible, we're
starting with a very limited amount of coins. Supported are all
Cryptonight coins including Monero, Aeon, Electroneum and more,
mining on the CPU only. OpenCL and CUDA support are at the top of our
priority list and will hopefully be ready with the next release.

Speaking of release, you can try the first alpha version of CoinStack right now.

<a href="#" class="button is-primary">Download for Linux</a>
<a href="#" class="button is-primary">Download for Windows</a>

After downloading CoinStack, you'll first want to run `coinstack miner --init` to generate
the configuration file. This will scan your hardware and generate a default configuration file
that looks like this:

```yml
donate: 5
coins:
  XMR:
    pool:
      url: stratum+tcp://xmr.coinfoundry.org:3032
      user: 46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr
      pass: x
cpus:
  "0":
    model: Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
    coin: XMR
    threads: 2
```

Unless you want to donate all your hashing power to us, you should at least change the pool user to
match your wallet address. Note that the donate value is ignored at the moment because donations are
not yet implemented, but you can simply remove it if you don't want to donate in the future.

After modifying the configuration to fit your needs, start the miner:

```
$ coinstack miner
19:55:03	info	miner started
19:56:03	info	CPU 0: 75.32 H/s
```

As you can see, the miner will automatically report your hashrate every minute.

But this is not all. CoinStack also includes a simple graphical user interface. On Windows, simply
double click the exe to launch it, on Linux you'll need to run `coinstack gui`. Note that this
requires webkit2gtk to be installed. Here's how it looks:


![](/images/2018/coinstack.png)

In case you're experiencing any issues, run CoinStack with the `--debug` flag and create
an issue [on our GitHub page](#) with the output.

We have big plans for CoinStack, but to make it happen, we need your help in any of the following
areas:

* **Programming**: Most of our code is written in Go, the crypto is implemented in C and the graphical
user interface uses JavaScript and HTML. If you're interested in contributing in any of these areas,
please get in touch (or just send use these delicious pull requests).

* **Donations**: We need to purchase a dual GPU mining rig for OpenCL/CUDA development and testing.
This will cost us around $450. If you can spare some Monero, please send them to `46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr`.
