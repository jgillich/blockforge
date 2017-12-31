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

	c := make(chan Object)
	topology.GetByType(ObjectTypePU, c)
	count := 0

	for o := range c {
		if o.TypeString() != "PU" {
			t.Fatalf("unexpected type '%s'", o.TypeString())
		}
		count++
	}

	if count != runtime.NumCPU() {
		t.Fatalf("unexpected core count '%v'", count)
	}
}
