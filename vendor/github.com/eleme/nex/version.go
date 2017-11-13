package nex

import (
	"fmt"
)

type version struct {
	Major, Minor, Patch int
}

var (
	// Version represents the version of this project.
	Version = version{0, 1, 0}
	// Build represents the build info of this project.
	Build = "unknown"
)

func (v version) String() string {
	return fmt.Sprintf("v%d.%d.%d Build: %s", v.Major, v.Minor, v.Patch, Build)
}
