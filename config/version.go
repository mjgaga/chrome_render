package config

import "fmt"

var (
	// Version should be updated by hand at each release
	Version = "0.0.0"

	//will be overwritten automatically by the build system
	ProjectName string = "chrome_render"
	GitCommit   string
	GoVersion   string
	BuildTime   string
)

// FullVersion formats the version to be printed
func FullVersion() string {
	return fmt.Sprintf("Version: %6s \nProject Name: %6s \nGit commit: %6s \nGo version: %6s \nBuild time: %6s \n",
		Version, ProjectName, GitCommit, GoVersion, BuildTime)
}
