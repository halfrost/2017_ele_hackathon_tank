package sdk

type api int

const (
	apiStats api = iota
	apiPorts
	apiRegistry
	apiApplication
	apiDependency
)

func (a api) String() string {
	if name, found := apiToName[a]; found {
		return name
	}
	return "Unknown API"
}

var (
	lowestVersionOfAPI = map[api]samVersion{
		apiStats:       assertNewVersionFromString("0.1"),
		apiPorts:       assertNewVersionFromString("0.2"),
		apiRegistry:    assertNewVersionFromString("0.2"),
		apiApplication: assertNewVersionFromString("0.3.2"),
		apiDependency:  assertNewVersionFromString("0.3.2"),
	}
	apiToName = map[api]string{
		apiStats:       "Stats",
		apiPorts:       "Ports",
		apiRegistry:    "Registry",
		apiApplication: "Application",
		apiDependency:  "Dependency",
	}
)
