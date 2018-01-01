package xmrstak

// #cgo CXXFLAGS: -std=c++11 -I${SRCDIR}/xmr-stak
// #cgo LDFLAGS: -L${SRCDIR}/xmr-stak/build/bin -lxmr-stak-backend -lxmr-stak-c -lssl -lcrypto -lhwloc
import "C"
