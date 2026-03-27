package version

var (
	// Version is the semantic version injected at build time (e.g., "250131.01.1")
	// Set via: go build -ldflags "-X dkhalife.com/tasks/core/internal/version.Version=1.0.0"
	Version = "dev"

	// BuildNumber is the CI build number injected at build time
	// Set via: go build -ldflags "-X dkhalife.com/tasks/core/internal/version.BuildNumber=123"
	BuildNumber = "0"

	// CommitHash is the git commit SHA injected at build time
	// Set via: go build -ldflags "-X dkhalife.com/tasks/core/internal/version.CommitHash=abc123"
	CommitHash = "unknown"
)

func GetVersion() string {
	if Version == "dev" {
		return "dev"
	}
	return Version
}

func GetFullVersion() string {
	if Version == "dev" {
		return "dev (unknown)"
	}
	commit := CommitHash
	if len(commit) > 7 {
		commit = commit[:7]
	}
	return Version + " (build " + BuildNumber + ", commit " + commit + ")"
}
