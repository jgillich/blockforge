package hwloc

// #include <hwloc.h>
// #cgo CFLAGS: -Wno-deprecated-declarations
// #cgo LDFLAGS: -lhwloc
import "C"

import (
	"errors"
	"unsafe"
)

var (
	errInternalHwlocError = errors.New("internal hwloc error")
)

type TopologyFlag uint64

const (
	TopologyFlagWholeSystem TopologyFlag = C.HWLOC_TOPOLOGY_FLAG_WHOLE_SYSTEM
	TopologyFlagThisSystem               = C.HWLOC_TOPOLOGY_FLAG_IS_THISSYSTEM
	TopologyFlagIODevices                = C.HWLOC_TOPOLOGY_FLAG_IO_DEVICES
	TopologyFlagIOBridges                = C.HWLOC_TOPOLOGY_FLAG_IO_BRIDGES
	TopologyFlagWholeIO                  = C.HWLOC_TOPOLOGY_FLAG_WHOLE_IO
	TopologyFlagICaches                  = C.HWLOC_TOPOLOGY_FLAG_ICACHES
)

type ObjectType int

const (
	ObjectTypeSystem    ObjectType = C.HWLOC_OBJ_SYSTEM
	ObjectTypeMachine              = C.HWLOC_OBJ_MACHINE
	ObjectTypeNumaNode             = C.HWLOC_OBJ_NUMANODE
	ObjectTypePackage              = C.HWLOC_OBJ_PACKAGE
	ObjectTypeCache                = C.HWLOC_OBJ_CACHE
	ObjectTypeCore                 = C.HWLOC_OBJ_CORE
	ObjectTypePU                   = C.HWLOC_OBJ_PU
	ObjectTypeGroup                = C.HWLOC_OBJ_GROUP
	ObjectTypeMisc                 = C.HWLOC_OBJ_MISC
	ObjectTypeBridge               = C.HWLOC_OBJ_BRIDGE
	ObjectTypePciDevice            = C.HWLOC_OBJ_PCI_DEVICE
	ObjectTypeOsDevice             = C.HWLOC_OBJ_OS_DEVICE
	ObjectTypeTypeMax              = C.HWLOC_OBJ_TYPE_MAX
)

// Topology represents the hardware layout of a machine.
type Topology interface {
	GetRootObj() Object
	GetNbobjsByType(ObjectType) int
	GetObjByType(ObjectType, int) Object
}

func NewTopology(flag TopologyFlag) (Topology, error) {
	t := &topology{}

	var r C.int

	r = C.hwloc_topology_init(&t.ptr)

	if r != 0 {
		return nil, errInternalHwlocError
	}

	C.hwloc_topology_set_flags(t.ptr, C.ulong(flag))

	r = C.hwloc_topology_load(t.ptr)

	if r != 0 {
		return nil, errInternalHwlocError
	}

	return t, nil
}

type topology struct {
	ptr C.hwloc_topology_t
}

func (t *topology) GetRootObj() Object {
	o := C.hwloc_get_root_obj(t.ptr)
	return &object{ptr: o}
}

func (t *topology) GetNbobjsByType(ot ObjectType) int {
	n := C.hwloc_get_nbobjs_by_type(t.ptr, C.hwloc_obj_type_t(ot))
	return int(n)
}

func (t *topology) GetObjByType(ot ObjectType, idx int) Object {
	o := C.hwloc_get_obj_by_type(t.ptr, C.hwloc_obj_type_t(ot), C.uint(idx))
	return &object{ptr: o}
}

type Object interface {
	Name() string
	Type() ObjectType
	TypeString() string
	InfoByName(string) string
}

type object struct {
	ptr C.hwloc_obj_t
}

func (o *object) Name() string {
	return C.GoString(o.ptr.name)
}

func (o *object) Type() ObjectType {
	return ObjectType(o.ptr._type)
}

func (o *object) TypeString() string {
	b := [64]C.char{}
	n := C.hwloc_obj_type_snprintf(&b[0], C.size_t(len(b)), o.ptr, 0)
	return C.GoStringN(&b[0], n)
}

func (o *object) InfoByName(name string) string {
	n := C.CString(name)
	c := C.hwloc_obj_get_info_by_name(o.ptr, n)
	C.free(unsafe.Pointer(n))
	return C.GoString(c)
}
