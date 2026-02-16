package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const (
	githubAPIURL = "https://api.github.com/repos/daylamtayari/Cierge/releases?per_page=20"
	cliTagPrefix = "cli/"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Retrieve program version",
	Run: func(cmd *cobra.Command, args []string) {
		version := getVersion()
		if version != "unknown" {
			version = "version " + version
		}

		fmt.Print(version)
	},
}

// Retrieve the version of the CLI
func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		version := info.Main.Version
		if version == "" || version == "(devel)" {
			version = "dev"
		}
		return version
	} else {
		return "unknown"
	}
}

// Queries the GitHub API to check if a newer CLI release is available
// Returns the latest version string and a boolean if an update is available
func checkForUpdate() (string, bool) {
	currentVersion := getVersion()
	if currentVersion == "dev" || currentVersion == "unknown" {
		return "", false
	}
	// Ensure version has "v" prefix for semver package
	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}
	// Validate current version
	if !semver.IsValid(currentVersion) {
		logger.Error().Str("version", currentVersion).Msg("Current version is not valid semver")
		return "", false
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create GitHub API request")
		return "", false
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Debug().Err(err).Msg("Failed to fetch releases from GitHub")
		return "", false
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		logger.Warn().Int("status", resp.StatusCode).Msg("GitHub API returned non-200 status")
		return "", false
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		logger.Error().Err(err).Msg("Failed to decode GitHub releases")
		return "", false
	}

	// Find the latest CLI release
	var latestVersion string
	var latestTag string

	for _, release := range releases {
		// Skip drafts and prereleases
		if release.Draft || release.Prerelease {
			continue
		}
		// Only check releases with CLI tag prefix
		if !strings.HasPrefix(release.TagName, cliTagPrefix) {
			continue
		}

		versionStr := strings.TrimPrefix(release.TagName, cliTagPrefix)
		if !strings.HasPrefix(versionStr, "v") {
			versionStr = "v" + versionStr
		}
		if !semver.IsValid(versionStr) {
			logger.Debug().Str("tag", release.TagName).Msg("Release tag is not valid semver")
			continue
		}

		// Compare versions (returns 1 if first > second, 0 if equal, -1 if first < second)
		if latestVersion == "" || semver.Compare(versionStr, latestVersion) > 0 {
			latestVersion = versionStr
			latestTag = release.TagName
		}
	}

	if latestVersion == "" {
		logger.Debug().Msg("No valid CLI releases found on GitHub")
		return "", false
	}

	if semver.Compare(latestVersion, currentVersion) > 0 {
		return latestTag, true
	}

	return "", false
}
