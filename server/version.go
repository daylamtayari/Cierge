package main

import "runtime/debug"

// Version is set at build time
var Version = "dev"

// getVersion returns the version, falling back to build info
func getVersion() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return Version
}
