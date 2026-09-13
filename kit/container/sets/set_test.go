package sets

import (
	"encoding/json"
	"testing"
)

func TestSetOperations(t *testing.T) {
	set := New("one", "two", "two")
	if !set.Has("one") || set.Has("missing") {
		t.Fatal("Has() returned an unexpected result")
	}
	if !set.HasAll("one", "two") || set.HasAll("one", "missing") {
		t.Fatal("HasAll() returned an unexpected result")
	}
	if !set.HasAny("missing", "two") || set.HasAny("missing", "other") {
		t.Fatal("HasAny() returned an unexpected result")
	}
	if len(set.ToSlice()) != 2 {
		t.Fatalf("ToSlice() length = %d; want 2", len(set.ToSlice()))
	}

	set.Delete("one").Insert("three")
	if set.Has("one") || !set.Has("three") {
		t.Fatal("Insert() or Delete() failed")
	}
	clone := set.Clone()
	clone.Insert("clone-only")
	if set.Has("clone-only") {
		t.Fatal("Clone() shares entries with the original")
	}

	set.Clear()
	if len(set.ToSlice()) != 0 {
		t.Fatal("Clear() did not empty the set")
	}
}

func TestSetJSON(t *testing.T) {
	set := New("one", "two")
	data, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Set[string]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.HasAll("one", "two") || len(decoded.ToSlice()) != 2 {
		t.Fatalf("JSON round trip = %#v", decoded.ToSlice())
	}
	if err := json.Unmarshal([]byte(`[1`), &decoded); err == nil {
		t.Fatal("UnmarshalJSON() accepted invalid JSON")
	}
}
