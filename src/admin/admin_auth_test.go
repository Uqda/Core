package admin

import (
	"testing"
)

func TestRequireAuth(t *testing.T) {
	t.Parallel()
	a := &AdminSocket{authToken: "secret-token"}
	t.Run("no token configured", func(t *testing.T) {
		t.Parallel()
		open := &AdminSocket{}
		if err := open.requireAuth(&AdminSocketRequest{Name: "list", Auth: ""}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("valid token", func(t *testing.T) {
		t.Parallel()
		if err := a.requireAuth(&AdminSocketRequest{Name: "list", Auth: "secret-token"}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("wrong token", func(t *testing.T) {
		t.Parallel()
		if err := a.requireAuth(&AdminSocketRequest{Name: "list", Auth: "wrong"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("missing token", func(t *testing.T) {
		t.Parallel()
		if err := a.requireAuth(&AdminSocketRequest{Name: "list", Auth: ""}); err == nil {
			t.Fatal("expected error")
		}
	})
}
