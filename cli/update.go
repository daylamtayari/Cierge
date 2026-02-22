package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const githubReleaseDownloadURL = "https://github.com/daylamtayari/Cierge/releases/download"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the CLI to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		latestTag, updateAvailable := checkForUpdate()
		if !updateAvailable {
			fmt.Println("Already on the latest version.")
			return
		}

		version := strings.TrimPrefix(latestTag, cliTagPrefix)
		logger.Info().Str("version", latestTag).Msg("Updating...")

		tmp, err := os.MkdirTemp("", "cierge-update-*")
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create temp directory")
		}
		defer os.RemoveAll(tmp)

		execPath, err := os.Executable()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to determine executable path")
		}
		execPath, err = filepath.EvalSymlinks(execPath)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to resolve executable path")
		}

		newBinary, err := downloadUpdate(tmp, version, latestTag)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to download update")
		}

		if err := os.Chmod(newBinary, 0755); err != nil {
			logger.Fatal().Err(err).Msg("Failed to set permissions on update")
		}

		if err := replaceExecutable(newBinary, execPath); err != nil {
			logger.Fatal().Err(err).Msg("Failed to install update")
		}

		fmt.Printf("Successfully updated to %s\n", latestTag)
	},
}

// downloadAsset downloads a file from url into dir, returning the local path
func downloadAsset(dir, url string) (string, error) {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	dst := filepath.Join(dir, filepath.Base(url))
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	return dst, nil
}
