package version

import (
	"fmt"
	"runtime"
)

var (
	// NAME .
	NAME = "Eru-Resrouce-Storage"
	// VERSION .
	VERSION = "unknown"
	// REVISION .
	REVISION = "HEAD"
	// BUILTAT .
	BUILTAT = "now"
)

// String .
func String() string {
	version := ""
	version += fmt.Sprintf("Version:        %s\n", VERSION)
	version += fmt.Sprintf("Git hash:       %s\n", REVISION)
	version += fmt.Sprintf("Built:          %s\n", BUILTAT)
	version += fmt.Sprintf("Golang version: %s\n", runtime.Version())
	version += fmt.Sprintf("OS/Arch:        %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return version
}
