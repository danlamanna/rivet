package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

var GitCommit string

const Version = "0.0.3"

var BuildDate = ""
var GoVersion = runtime.Version()
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

type GithubRelease struct {
	Version string `json:"name"`
}

func IsNewVersionAvailable() (string, error) {
	client := &http.Client{
		Timeout: time.Millisecond * 500,
	}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/danlamanna/rivet/releases", nil)

	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code %d when fetching releases", resp.StatusCode)
	}

	releases := make([]GithubRelease, 1)
	if err = json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}

	currentVersion, err := version.NewVersion(Version)
	latestVersion, err := version.NewVersion(strings.TrimPrefix(releases[0].Version, "v"))

	if err != nil {
		return "", err
	}

	if latestVersion.GreaterThan(currentVersion) {
		return latestVersion.String(), nil
	}

	return "", nil
}
