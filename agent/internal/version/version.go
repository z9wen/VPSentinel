package version

// Release builds inject these values through -ldflags. Local builds retain
// dev/unknown so operators can distinguish them from published artifacts.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
