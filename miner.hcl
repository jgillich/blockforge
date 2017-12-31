donate = 1

cpu "0" {
  name = "Intel i5"
  threads = 2
  coin = "monero"
}

gpu "0" {
  name = "Radeon RX 560"
  coin = "monero"
}

coin "monero" {
  pool {
    address = "stratum+tcp://xmr.poolmining.org:3032",
    user = "46DTAEGoGgc575EK7rLmPZFgbXTXjNzqrT4fjtCxBFZSQr5ScJFHyEScZ8WaPCEsedEFFLma6tpLwdCuyqe6UYpzK1h3TBr",
    password = "x",
  }
}
