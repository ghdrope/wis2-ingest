package version

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// These variables are set at build time using -ldflags.
// Default values are used when running locally (e.g. `go run`).
var (
	Version   = "development"
	GitCommit = "none"
	BuildDate = "unknown"

	GoVersion = runtime.Version()
	Platform  = runtime.GOOS + "/" + runtime.GOARCH
)

// Info represents structured version information.
type Info struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
	Platform  string
}

// Get returns the current version information as a struct.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  Platform,
	}
}

// String returns a multi-line version string.
func (i Info) String() string {
	v := reflect.ValueOf(i)
	t := v.Type()
	var sb strings.Builder
	for j := 0; j < v.NumField(); j++ {
		name := t.Field(j).Name
		value := v.Field(j).Interface()
		fmt.Fprintf(&sb, "%-10s: %v", name, value)
		if j < v.NumField()-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// Short returns only the version string.
func (i Info) Short() string {
	return fmt.Sprintf("Version:    %s", i.Version)
}

// String returns full version info directly (helper).
func String() string {
	return Get().String()
}

// Short returns only version (helper).
func Short() string {
	return Get().Short()
}
