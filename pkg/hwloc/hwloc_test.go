package hwloc

import (
	"runtime"
	"testing"
)

func TestCoreCount(t *testing.T) {
	topology, err := NewTopology(TopologyFlagWholeSystem)
	if err != nil {
		t.Error(err)
	}

	width := topology.GetNbobjsByType(ObjectTypePU)
	count := 0

	for i := 0; i < width; i++ {
		o := topology.GetObjByType(ObjectTypePU, uint(i))
		if o.TypeString() != "PU" {
			t.Fatalf("unexpected type '%s'", o.TypeString())
		}
		count++
	}

	if count != runtime.NumCPU() {
		t.Fatalf("unexpected core count '%v'", count)
	}
}
