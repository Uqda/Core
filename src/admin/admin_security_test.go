package admin

import (
	"strings"
	"testing"
)

// Security-focused tests for admin API authentication (shared secret).

func TestRequireAuth_RejectsWrongSecretVariants(t *testing.T) {
	t.Parallel()
	const good = "correct-horse-battery-staple"
	a := &AdminSocket{authToken: good}

	cases := []struct {
		name string
		auth string
	}{
		{"empty when required", ""},
		{"single char off", good[:len(good)-1] + "x"},
		{"prefix only", good[:3]},
		{"suffix only", good[3:]},
		{"case change", strings.ToUpper(good)},
		{"extra suffix", good + "x"},
		{"leading space", " " + good},
		{"null byte injection", good + "\x00"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := a.requireAuth(&AdminSocketRequest{Name: "list", Auth: tc.auth}); err == nil {
				t.Fatalf("expected auth failure for variant %q", tc.name)
			}
		})
	}
}

func TestRequireAuth_AcceptsExactMatch(t *testing.T) {
	t.Parallel()
	const tok = "a\x00b" // token may contain arbitrary bytes
	a := &AdminSocket{authToken: tok}
	if err := a.requireAuth(&AdminSocketRequest{Name: "list", Auth: tok}); err != nil {
		t.Fatal(err)
	}
}

func TestRequireAuth_UnicodeToken(t *testing.T) {
	t.Parallel()
	tok := "كلمة-سر-🔐"
	a := &AdminSocket{authToken: tok}
	if err := a.requireAuth(&AdminSocketRequest{Name: "getSelf", Auth: tok}); err != nil {
		t.Fatal(err)
	}
	if err := a.requireAuth(&AdminSocketRequest{Name: "getSelf", Auth: tok + "x"}); err == nil {
		t.Fatal("expected failure")
	}
}
