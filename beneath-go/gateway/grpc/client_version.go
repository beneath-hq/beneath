package grpc

import version "github.com/hashicorp/go-version"

type clientVersionSpec struct {
	RecommendedVersion *version.Version
	WarningVersion     *version.Version
	DeprecatedVersion  *version.Version
}

func (s clientVersionSpec) IsZero() bool {
	return s.DeprecatedVersion == nil && s.WarningVersion == nil && s.RecommendedVersion == nil
}

func newClientVersionSpec(recommended, warning, deprecated string) clientVersionSpec {
	return clientVersionSpec{
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

var clientSpecs = map[string]clientVersionSpec{
	"beneath-python": newClientVersionSpec("1.0.2", "1.0.1", "0.0.1"),
}
