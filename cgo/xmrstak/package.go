package cgo

// #cgo LDFLAGS: -L${SRCDIR}/cgo/xmr-stak/bin -lxmr-stak-backend -lxmr-stak-c -lssl -lcrypto -lhwloc
import "C"
