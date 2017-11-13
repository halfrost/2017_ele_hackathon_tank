package sdk

import (
	"fmt"
	"strconv"
	"strings"
)

type samVersion []int

// String returns the version separated by dot.
func (sv samVersion) String() string {
	var version = ""
	for _, subVer := range sv {
		if len(version) > 0 {
			version += "."
		}
		version += fmt.Sprint(subVer)
	}
	return version
}

// Less returns true if less than given version.
func (sv samVersion) Less(other samVersion) bool {
	var minSize = len(sv)
	if minSize > len(other) {
		minSize = len(other)
	}

	for i := 0; i < minSize; i++ {
		if sv[i] == other[i] {
			continue
		}
		return sv[i] < other[i]
	}

	return len(sv) < len(other)
}

func newVersionFromString(ver string) (samVersion, error) {
	rawVersion := strings.Split(ver, ".")
	var samVersion = make(samVersion, len(rawVersion))

	for i, raw := range rawVersion {
		sub, err := strconv.Atoi(raw)
		if err != nil {
			return samVersion, err
		}
		samVersion[i] = sub
	}

	return samVersion, nil
}

// assertNewVersionFromString always returns an valid samVersion.
// It panic if given version is invalid.
func assertNewVersionFromString(ver string) samVersion {
	v, err := newVersionFromString(ver)
	if err != nil {
		panic(err)
	}
	return v
}
