package ethminer

// #cgo CXXFLAGS: -std=c++11 -I${SRCDIR}/ethminer -DETH_STRATUM
// #cgo LDFLAGS: -lboost_system -ljsoncpp
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libstratum -lethstratum
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libethcore -lethcore
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libethash -lethash
// #cgo LDFLAGS: -L${SRCDIR}/ethminer/build/libdevcore -ldevcore
import "C"
