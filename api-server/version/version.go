package version

import "fmt"

var (
	Version   = "0.0.1"
	GitCommit = "HEAD"
	BuildDate = "1970-01-01T00:00:00Z"
)

func Print() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("GitCommit: %s\n", GitCommit)
	fmt.Printf("BuildDate: %s\n", BuildDate)
}
