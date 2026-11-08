package core

import (
	"fmt"
	"net/url"

	"golang.org/x/crypto/blake2b"

	"github.com/Uqda/Core/src/config"
)

// linkPassword returns the link password from a peer or listener URI. It
// accepts ?password=TEXT or ?password_file=PATH, which keeps the secret out
// of the configuration file. It returns nil when neither is set.
func linkPassword(q url.Values) ([]byte, error) {
	p, file := q.Get("password"), q.Get("password_file")
	switch {
	case p != "" && file != "":
		return nil, fmt.Errorf("%w: use password or password_file, not both", ErrLinkPasswordInvalid)
	case file != "":
		secret, err := config.ReadSecretFile(file)
		if err != nil {
			return nil, fmt.Errorf("%w: password_file: %v", ErrLinkPasswordInvalid, err)
		}
		p = secret
	case p == "":
		return nil, nil
	}
	if len(p) > blake2b.Size {
		return nil, fmt.Errorf("%w: at most %d bytes", ErrLinkPasswordInvalid, blake2b.Size)
	}
	return []byte(p), nil
}
