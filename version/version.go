package version

import (
	"strings"

	"github.com/tcnksm/go-latest"
)

var (
	// BuildTime is the time for the current build
	BuildTime string
	// BuildVersion is the version for the current build
	BuildVersion string
)

// CheckForUpdate checks GitHub for a newer version
func CheckForUpdate() string {
	if !strings.HasPrefix(BuildVersion, "v") {
		return ""
	}
	tag := &latest.GithubTag{
		Owner:             "screepers",
		Repository:        "screeps-launcher",
		FixVersionStrFunc: latest.DeleteFrontV(),
	}
	res, err := latest.Check(tag, BuildVersion)
	if err == nil && res.Outdated {
		return res.Current
	}
	return ""
}
