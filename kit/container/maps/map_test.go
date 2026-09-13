package maps

import (
	"encoding/json"
	"testing"
)

func TestMapOperations(t *testing.T) {
	m := New(map[string]int{"one": 1})

	if value, ok := m.Load("one"); !ok || value != 1 {
		t.Fatalf("Load() = %d, %t; want 1, true", value, ok)
	}
	if value, ok := m.Load("missing"); ok || value != 0 {
		t.Fatalf("Load() for missing key = %d, %t; want 0, false", value, ok)
	}
	if value, loaded := m.LoadOrStore("one", 2); !loaded || value != 1 {
		t.Fatalf("LoadOrStore() existing = %d, %t; want 1, true", value, loaded)
	}
	if value, loaded := m.LoadOrStore("two", 2); loaded || value != 2 {
		t.Fatalf("LoadOrStore() new = %d, %t; want 2, false", value, loaded)
	}
	if !m.CompareAndSwap("one", 1, 3) || m.CompareAndSwap("one", 1, 4) {
		t.Fatal("CompareAndSwap() did not enforce the old value")
	}
	if !m.CompareAndDelete("one", 3) || m.CompareAndDelete("one", 3) {
		t.Fatal("CompareAndDelete() did not enforce the current value")
	}

	m.Store("three", 3)
	if value, loaded := m.LoadAndDelete("three"); !loaded || value != 3 {
		t.Fatalf("LoadAndDelete() = %d, %t; want 3, true", value, loaded)
	}
	if value, loaded := m.Swap("two", 4); !loaded || value != 2 {
		t.Fatalf("Swap() = %v, %t; want 2, true", value, loaded)
	}
	if value, loaded := m.Swap("four", 4); loaded || value != nil {
		t.Fatalf("Swap() for new key = %v, %t; want nil, false", value, loaded)
	}

	seen := m.ToMap()
	if len(seen) != 2 || seen["two"] != 4 || seen["four"] != 4 {
		t.Fatalf("ToMap() = %#v", seen)
	}
	clone := m.Clone()
	clone.Store("clone-only", 5)
	if _, ok := m.Load("clone-only"); ok {
		t.Fatal("Clone() shares entries with the original")
	}
	m.Clear()
	if len(m.ToMap()) != 0 {
		t.Fatal("Clear() did not empty the map")
	}
}

func TestMapJSON(t *testing.T) {
	m := New(map[string]int{"one": 1, "two": 2})
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Map[string, int]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.ToMap()
	if len(got) != 2 || got["one"] != 1 || got["two"] != 2 {
		t.Fatalf("JSON round trip = %#v", got)
	}
	if err := json.Unmarshal([]byte(`{"broken":`), &decoded); err == nil {
		t.Fatal("UnmarshalJSON() accepted invalid JSON")
	}
}
