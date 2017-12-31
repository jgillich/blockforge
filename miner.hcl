donate = 1

cpu "Intel i5" {
  threads = 2
  coin = "monero"
}

gpu "RX 560" {
  index = 0
  coin = "monero"
}

coin "monero" {
  pool {
    url = "stratum+tcp://xmr.poolmining.org:3032",
    user = "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
    pass = "x",
  }
}
