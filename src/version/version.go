// Package version defines Uqda's product identity independently of wire versions.
package version

import (
	_ "embed" // Required by go:embed.
	"strings"
)

//go:embed VERSION
var releaseVersion string

// BuildName is the stable implementation identifier used by diagnostics.
func BuildName() string { return "uqda" }

// BuildVersion returns the canonical machine/package version.
func BuildVersion() string { return strings.TrimSpace(releaseVersion) }

// ProductName formats the annual version for public display.
func ProductName() string {
	v := strings.SplitN(BuildVersion(), "-", 2)
	name := "Uqda " + strings.TrimSuffix(v[0], ".0")
	if len(v) == 2 {
		if beta, ok := strings.CutPrefix(v[1], "beta."); ok {
			name += " Beta " + beta
		} else {
			name += " " + v[1]
		}
	}
	return name
}

// DisplayName returns the human-facing Core release name.
func DisplayName() string { return strings.Replace(ProductName(), "Uqda ", "Uqda Core ", 1) }
