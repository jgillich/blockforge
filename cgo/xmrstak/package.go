package xmrstak

import _ "gitlab.com/jgillich/autominer/cgo/cl"

// #cgo CXXFLAGS: -std=c++11 -I${SRCDIR}/xmr-stak
// #cgo LDFLAGS: -L${SRCDIR}/xmr-stak/build/bin -lxmr-stak-backend -lxmr-stak-c -lssl -lcrypto -lhwloc -ldl
import "C"
