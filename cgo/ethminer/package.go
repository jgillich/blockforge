package ethminer

// #cgo CXXFLAGS: -I${SRCDIR}/cgo/ethminer/ethminer/libethcore
// #cgo LDFLAGS: -L${SRCDIR}/cgo/ethminer/ethminer -L${SRCDIR}/cgo/ethminer/ethminer/libethcore -lboost
import "C"
