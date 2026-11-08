package core

import (
	"net/url"
	"strings"
)

// secretQueryKeys are URI query parameters whose values must never be logged.
var secretQueryKeys = []string{"password"}

// RedactURI returns raw with link passwords and URL user credentials
// replaced, so peer and listener URIs can be logged safely. If raw cannot be
// parsed, everything after the scheme and host is dropped.
func RedactURI(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		if i := strings.IndexAny(raw, "?#"); i >= 0 {
			return raw[:i] + "?…"
		}
		return raw
	}
	return redactURL(u)
}

func redactURL(u *url.URL) string {
	c := *u
	if c.User != nil {
		if _, has := c.User.Password(); has {
			c.User = url.UserPassword(c.User.Username(), "REDACTED")
		}
	}
	if c.RawQuery != "" {
		q := c.Query()
		changed := false
		for _, k := range secretQueryKeys {
			if q.Has(k) {
				q.Set(k, "REDACTED")
				changed = true
			}
		}
		if changed {
			c.RawQuery = q.Encode()
		}
	}
	return c.String()
}
