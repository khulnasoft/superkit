package button

import (
	"testing"

	"github.com/a-h/templ"
)

func TestButtonVariants(t *testing.T) {
	tests := []struct {
		name     string
		attrs    string
		contains string
	}{
		{name: "default", attrs: New()["class"].(string), contains: buttonVariantPrimary},
		{name: "primary", attrs: Primary()["class"].(string), contains: buttonVariantPrimary},
		{name: "outline", attrs: Outline()["class"].(string), contains: buttonVariantOutline},
		{name: "secondary", attrs: Secondary()["class"].(string), contains: buttonVariantSecondary},
		{name: "destructive", attrs: Destructive()["class"].(string), contains: buttonVariantDestructive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attrs != buttonBaseClass+" "+tt.contains {
				t.Fatalf("class = %q; want base and %q", tt.attrs, tt.contains)
			}
		})
	}
}

func TestButtonCustomOptions(t *testing.T) {
	attrs := Outline(func(attrs *templ.Attributes) {
		(*attrs)["data-test"] = "button"
	})
	if got := attrs["data-test"]; got != "button" {
		t.Fatalf("custom option = %q; want button", got)
	}
}
