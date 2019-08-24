package version

import (
	"fmt"
	"runtime"
)

var GitCommit string

const Version = "0.0.3"

var BuildDate = ""
var GoVersion = runtime.Version()
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
