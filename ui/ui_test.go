package ui

import (
	"testing"

	"github.com/a-h/templ"
)

func TestCreateAttrsAndClass(t *testing.T) {
	attrs := CreateAttrs("base", "default", Class("extra"))
	if got := attrs["class"]; got != "base default extra" {
		t.Fatalf("class = %q; want %q", got, "base default extra")
	}

	var opts []func(*templ.Attributes)
	attrs = CreateAttrs("base", "", opts...)
	if got := attrs["class"]; got != "base" {
		t.Fatalf("class = %q; want %q", got, "base")
	}
}

func TestMerge(t *testing.T) {
	if got := Merge("first", "second"); got != "first second" {
		t.Fatalf("Merge() = %q; want %q", got, "first second")
	}
	if got := Merge("", "second"); got != "second" {
		t.Fatalf("Merge() = %q; want %q", got, "second")
	}
	if got := Merge("first", ""); got != "first" {
		t.Fatalf("Merge() = %q; want %q", got, "first")
	}
}

func TestClassOptionMutatesAttributes(t *testing.T) {
	attrs := CreateAttrs("base", "default")
	Class("extra")(&attrs)
	if got := attrs["class"]; got != "base default extra" {
		t.Fatalf("class = %q; want %q", got, "base default extra")
	}
}

func TestClassOptionHandlesEmptyAttributes(t *testing.T) {
	attrs := templ.Attributes{}
	Class("extra")(&attrs)
	if got := attrs["class"]; got != "extra" {
		t.Fatalf("class = %q; want %q", got, "extra")
	}
}
