package core

import (
	"strings"
	"testing"
)

func TestRedactURI(t *testing.T) {
	cases := map[string]string{
		"tls://[::]:443?password=hunter2":              "tls://[::]:443?password=REDACTED",
		"tcp://1.2.3.4:5?key=abc&password=s3cret":      "tcp://1.2.3.4:5?key=abc&password=REDACTED",
		"socks://user:pass@proxy:1080/peer:443":        "socks://user:REDACTED@proxy:1080/peer:443",
		"tls://host:443?password_file=/run/creds/peer": "tls://host:443?password_file=/run/creds/peer",
		"tcp://host:1":            "tcp://host:1",
		"%zz://bad?password=leak": "%zz://bad?…",
	}
	for in, want := range cases {
		got := RedactURI(in)
		if got != want {
			t.Errorf("RedactURI(%q) = %q, want %q", in, got, want)
		}
		if strings.Contains(got, "hunter2") || strings.Contains(got, "s3cret") || strings.Contains(got, "leak") || strings.Contains(got, ":pass@") {
			t.Errorf("RedactURI(%q) leaked a secret: %q", in, got)
		}
	}
}
