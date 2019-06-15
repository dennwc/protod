package protod

import "testing"

func TestGetType(t *testing.T) {
	w := &writer{pkg: "a.b", path: []string{"c"}}
	for _, c := range [][2]string{
		{"a.b.c.d", "d"},
		{"a.b.e", "e"},
		{"a.f.e", "a.f.e"},
	} {
		t.Run(c[0], func(t *testing.T) {
			if got := w.getType(c[0]); got != c[1] {
				t.Fatalf("unexpected type: %q vs %q", got, c[1])
			}
		})
	}
}
