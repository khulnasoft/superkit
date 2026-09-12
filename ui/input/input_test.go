package input

import "testing"

func TestInput(t *testing.T) {
	attrs := Input()
	if got := attrs["class"]; got != defaultInputClass+" " {
		t.Fatalf("class = %q; want %q", got, defaultInputClass+" ")
	}
}
