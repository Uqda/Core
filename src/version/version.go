package version

// BuildName and BuildVersion are set at link time, for example:
//
//	-X github.com/Uqda/Core/src/version.BuildName=uqda
//	-X github.com/Uqda/Core/src/version.BuildVersion=0.5.13
var BuildName string
var BuildVersion string

// Name returns the injected build name, or "unknown" if unset.
func Name() string {
	if BuildName == "" {
		return "unknown"
	}
	return BuildName
}

// SemVer returns the injected build version, or "unknown" if unset.
func SemVer() string {
	if BuildVersion == "" {
		return "unknown"
	}
	return BuildVersion
}
