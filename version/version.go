package version

// Version related constants.
const (
	Major = "2"
	Minor = "2"
	Fix   = "2"
)

var (
	// Version is the current version of AnychainDB platform.
	Version = "2.2.2"

	// GitCommit is the current HEAD set using ldflags.
	GitCommit string
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
