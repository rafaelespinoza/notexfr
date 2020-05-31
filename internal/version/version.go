package version

import "fmt"

// These are pieces of version metadata that can be set through -ldflags.
var (
	BranchName string
	BuildTime  string
	CommitHash string
	GoOSArch   string
	GoVersion  string
	ReleaseTag string
)

// Show outputs metadata about the build. Values are set with go build -ldflags.
func Show() {
	fmt.Println("notexfr build metadata")
	fmt.Printf("\tBranchName: 	%s\n", BranchName)
	fmt.Printf("\tBuildTime:  	%s\n", BuildTime)
	fmt.Printf("\tCommitHash: 	%s\n", CommitHash)
	fmt.Printf("\tGoOSArch: 	%s\n", GoOSArch)
	fmt.Printf("\tGoVersion: 	%s\n", GoVersion)
	fmt.Printf("\tReleaseTag: 	%s\n", ReleaseTag)
}
