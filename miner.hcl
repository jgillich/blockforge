donate = 1

cpu "Intel i5" {
  threads = 1
  coin = "xmr"
}

gpu "RX 560" {
  index = 0
  coin = "eth"
}

coin "xmr" {
  pool {
    url = "stratum+tcp://xmr.poolmining.org:3032",
    user = "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
    pass = "x",
  }
}

coin "eth" {
  pool {
    url = "tcp://eu1.ethermine.org:4444"
    user = "0x25ae2cbddE36CfC9D959a4d1f76964EaE7517748"
    pass = "x"
  }
}