package version

// Release builds inject these values with -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
