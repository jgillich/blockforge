package ethminer

import _ "gitlab.com/jgillich/autominer/cgo/cl"

// #cgo CXXFLAGS: -std=c++11 -I${SRCDIR}/ethminer -I${SRCDIR}/ethminer/build -DETH_STRATUM -DETH_ETHASHCL
// #cgo LDFLAGS: -lboost_system -ljsoncpp -lOpenCL -ldl
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libethash-cl -lethash-cl
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libhwmon -lhwmon
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libstratum -lethstratum
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libethcore -lethcore
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libethash -lethash
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libdevcore -ldevcore
import "C"
