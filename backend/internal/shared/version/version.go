// Package version provides version information for the Solobueno ERP backend.
package version

// Version is the current version of the application.
const Version = "0.0.1"

// AppName is the name of the application.
const AppName = "Solobueno ERP"

// Info returns formatted version information.
func Info() string {
	return AppName + " v" + Version
}

// IsPreRelease returns true if the version is a pre-release version.
func IsPreRelease() bool {
	return Version[0] == '0'
}
