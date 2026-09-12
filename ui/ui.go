package ui

import (
	"strings"

	"github.com/a-h/templ"
)

func CreateAttrs(baseClass string, defaultClass string, opts ...func(*templ.Attributes)) templ.Attributes {
	attrs := templ.Attributes{
		"class": Merge(baseClass, defaultClass),
	}
	for _, o := range opts {
		o(&attrs)
	}
	return attrs
}

func Merge(a, b string) string {
	parts := []string{}
	for _, part := range []string{a, b} {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, " ")
}

func Class(class string) func(*templ.Attributes) {
	return func(attrs *templ.Attributes) {
		current, _ := (*attrs)["class"].(string)
		(*attrs)["class"] = Merge(current, class)
	}
}
