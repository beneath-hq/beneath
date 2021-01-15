package clientversion

import "github.com/hashicorp/go-version"

// Spec represents the deprecated, deprecating and recommended versions of a client
type Spec struct {
	RecommendedVersion *version.Version
	WarningVersion     *version.Version
	DeprecatedVersion  *version.Version
}

// IsZero returns true if the spec is uninitialized
func (s Spec) IsZero() bool {
	return s.DeprecatedVersion == nil && s.WarningVersion == nil && s.RecommendedVersion == nil
}

func newSpec(recommended, warning, deprecated string) Spec {
	return Spec{
		RecommendedVersion: newVersionOrPanic(recommended),
		WarningVersion:     newVersionOrPanic(warning),
		DeprecatedVersion:  newVersionOrPanic(deprecated),
	}
}

func newVersionOrPanic(str string) *version.Version {
	v, err := version.NewVersion(str)
	if err != nil {
		panic(err)
	}
	return v
}

// Specs contains a Spec for each supported client library
var Specs = map[string]Spec{
	"beneath-js":     newSpec("1.1.0", "1.0.4", "1.0.3"),
	"beneath-python": newSpec("1.2.11", "1.2.0", "1.1.3"),
}
