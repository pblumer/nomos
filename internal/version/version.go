package version

import "runtime"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	Dirty   = "unknown"
	BuiltBy = "source"
)

type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Dirty   string `json:"dirty"`
	BuiltBy string `json:"builtBy"`
	Go      string `json:"go"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

func Get() Info {
	return Info{
		Name:    "nomos",
		Version: Version,
		Commit:  Commit,
		Date:    Date,
		Dirty:   Dirty,
		BuiltBy: BuiltBy,
		Go:      runtime.Version(),
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}
}

func (i Info) OSArch() string {
	return i.OS + "/" + i.Arch
}
