package version

import "testing"

func TestProductNameFor(t *testing.T) {
	for _, tt := range []struct {
		machine string
		want    string
	}{
		{"26.0-beta.1", "Uqda 26 Beta 1"},
		{"26.0.0", "Uqda 26"},
		{"26.0.1", "Uqda 26.0.1"},
		{"26.1.0", "Uqda 26.1"},
		{"26.1", "Uqda 26.1"},
	} {
		if got := productNameFor(tt.machine); got != tt.want {
			t.Errorf("productNameFor(%q) = %q, want %q", tt.machine, got, tt.want)
		}
	}
}

func TestDisplayNameFor(t *testing.T) {
	for in, want := range map[string]string{
		"26.1.0":      "Uqda Core 26.1",
		"26.0.5":      "Uqda Core 26.0.5",
		"27.0.0":      "Uqda Core 27",
		"26.1-beta.1": "Uqda Core 26.1 Beta 1",
	} {
		if got := DisplayNameFor(in); got != want {
			t.Errorf("DisplayNameFor(%q) = %q, want %q", in, got, want)
		}
	}
}
