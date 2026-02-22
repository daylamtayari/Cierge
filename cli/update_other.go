//go:build !darwin

package main

import (
	"fmt"
	"os"
	"runtime"
)

// downloadUpdate downloads the raw binary for the current platform,
// returning the path to the downloaded file.
func downloadUpdate(dir, version, tag string) (string, error) {
	assetName := fmt.Sprintf("cierge-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}
	url := fmt.Sprintf("%s/%s/%s", githubReleaseDownloadURL, tag, assetName)

	logger.Info().Str("asset", assetName).Msg("Downloading...")
	return downloadAsset(dir, url)
}

// replaceExecutable replaces dst with src using an atomic rename.
func replaceExecutable(src, dst string) error {
	return os.Rename(src, dst)
}
