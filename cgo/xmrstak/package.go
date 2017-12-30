package xmrstak

// #cgo CXXFLAGS: -std=c++11
// #cgo LDFLAGS: -L${SRCDIR}/xmr-stak/bin -lxmr-stak-backend -lxmr-stak-c -lssl -lcrypto -lhwloc
import "C"
