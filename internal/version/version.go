package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Current represents the current version of the application
const (
	Current = "0.1.3"
	Repo    = "sametcn99/golter"
)

type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdates checks GitHub for the latest release
// Returns latest version string, release URL, true if update available, and error if any
func CheckForUpdates() (string, string, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", Repo)
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", false, fmt.Errorf("API responded with status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", false, err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Current, "v")

	if isNewer(current, latest) {
		return latest, release.HTMLURL, true, nil
	}

	return latest, "", false, nil
}

// isNewer returns true if v2 is newer than v1
// Assumes semantic versioning (x.y.z)
func isNewer(v1, v2 string) bool {
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")

	len1 := len(p1)
	len2 := len(p2)
	maxLen := len1
	if len2 > maxLen {
		maxLen = len2
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len1 {
			n1, _ = strconv.Atoi(p1[i])
		}
		if i < len2 {
			n2, _ = strconv.Atoi(p2[i])
		}

		if n2 > n1 {
			return true
		}
		if n1 > n2 {
			return false
		}
	}

	return false
}
