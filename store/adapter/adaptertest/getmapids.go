package adaptertest

import (
	"fmt"
	"testing"

	. "github.com/stratumn/go/segment/segmenttest"
	. "github.com/stratumn/go/store/adapter"
)

// Tests what happens when you get all the map IDs.
func TestGetMapIDsAll(t *testing.T, a Adapter) {
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			s := RandomSegment()
			s.Link.Meta["mapId"] = fmt.Sprintf("map%d", i)
			a.SaveSegment(s)
		}
	}

	slice, err := a.GetMapIDs(&Pagination{})

	if err != nil {
		t.Fatal(err)
	}

	if len(slice) != 10 {
		t.Fatal("expected map length to be 10")
	}

	for i := 0; i < 10; i++ {
		if !ContainsString(slice, fmt.Sprintf("map%d", i)) {
			t.Fatal("missing map ID")
		}
	}
}

// Tests what happens when you get map IDs with pagination.
func TestGetMapIDsPagination(t *testing.T, a Adapter) {
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			segment := RandomSegment()
			segment.Link.Meta["mapId"] = fmt.Sprintf("map%d", i)
			a.SaveSegment(segment)
		}
	}

	slice, err := a.GetMapIDs(&Pagination{3, 5})

	if err != nil {
		t.Fatal(err)
	}

	if len(slice) != 5 {
		t.Fatal("expected map length to be 5")
	}
}

// Tests what happens when you should get no map IDs.
func TestGetMapIDsEmpty(t *testing.T, a Adapter) {
	slice, err := a.GetMapIDs(&Pagination{100000, 5})

	if err != nil {
		t.Fatal(err)
	}

	if len(slice) != 0 {
		t.Fatal("expected map length to be 0")
	}
}
