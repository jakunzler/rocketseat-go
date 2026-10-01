package buildinfo

// Version and Commit are set with -ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
)
