package slices

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSliceOperations(t *testing.T) {
	list := New(1, 2, 3)
	if list.Len() != 3 {
		t.Fatalf("Len() = %d; want 3", list.Len())
	}
	if value, ok := list.Get(-1); ok || value != 0 {
		t.Fatalf("Get(-1) = %d, %t; want 0, false", value, ok)
	}
	if value, ok := list.Get(1); !ok || value != 2 {
		t.Fatalf("Get(1) = %d, %t; want 2, true", value, ok)
	}
	if !list.Set(1, 4) || !list.Set(2, 5) {
		t.Fatal("Set() returned an unexpected result")
	}
	if list.Append().Len() != 3 || list.Append(6).Len() != 4 {
		t.Fatal("Append() changed the list unexpectedly")
	}
	if !reflect.DeepEqual(list.Slice(1, 3), []int{4, 5}) ||
		!reflect.DeepEqual(list.SliceStart(2), []int{5, 6}) ||
		!reflect.DeepEqual(list.SliceEnd(2), []int{1, 4}) {
		t.Fatal("Slice helpers returned unexpected values")
	}

	seen := make([]int, 0)
	list.Range(func(index, item int) bool {
		seen = append(seen, index+item)
		return item != 4
	})
	if !reflect.DeepEqual(seen, []int{1, 5}) {
		t.Fatalf("Range() = %#v; want [1 5]", seen)
	}
	if value, ok := list.RemoveAt(0); !ok || value != 1 {
		t.Fatalf("RemoveAt() = %d, %t; want 1, true", value, ok)
	}
	if _, ok := list.RemoveAt(99); ok {
		t.Fatal("RemoveAt() accepted an invalid index")
	}

	clone := list.Clone()
	clone.Set(0, 99)
	if value, _ := list.Get(0); value == 99 {
		t.Fatal("Clone() shares backing data with the original")
	}
	copyOfList := list.ToSlice()
	copyOfList[0] = 100
	if value, _ := list.Get(0); value == 100 {
		t.Fatal("ToSlice() returned the backing data")
	}

	list.Clear()
	if list.Len() != 0 {
		t.Fatal("Clear() did not empty the list")
	}
}

func TestSliceJSON(t *testing.T) {
	list := New("one", "two")
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Slice[string]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.ToSlice(), []string{"one", "two"}) {
		t.Fatalf("JSON round trip = %#v", decoded.ToSlice())
	}
	if err := json.Unmarshal([]byte(`[1`), &decoded); err == nil {
		t.Fatal("UnmarshalJSON() accepted invalid JSON")
	}
}
