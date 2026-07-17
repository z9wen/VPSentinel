package version

// 这些变量会在发布构建时通过 -ldflags 注入。
// 本地直接 go run 时保留 dev/unknown，方便区分开发版本。
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
