//go:build darwin

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// downloadUpdate downloads the macOS .pkg for the current arch and extracts
// the binary from it, returning the path to the extracted binary.
func downloadUpdate(dir, version, tag string) (string, error) {
	assetName := fmt.Sprintf("cierge-%s-darwin-%s.pkg", version, runtime.GOARCH)
	url := fmt.Sprintf("%s/%s/%s", githubReleaseDownloadURL, tag, assetName)

	logger.Info().Str("asset", assetName).Msg("Downloading update")
	pkgPath, err := downloadAsset(dir, url)
	if err != nil {
		return "", err
	}

	return extractBinaryFromPkg(pkgPath, dir)
}

// extractBinaryFromPkg extracts the cierge binary from a .pkg installer.
// A .pkg is a XAR archive containing a gzipped CPIO payload.
func extractBinaryFromPkg(pkgPath, dir string) (string, error) {
	// Extract the XAR archive into dir, producing a Payload file
	xarCmd := exec.Command("xar", "-xf", pkgPath)
	xarCmd.Dir = dir
	if out, err := xarCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("xar extraction failed: %w\n%s", err, out)
	}

	// Extract the gzipped CPIO payload into a subdirectory
	extractDir := filepath.Join(dir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", err
	}

	payloadPath := filepath.Join(dir, "Payload")
	cpioCmd := exec.Command("sh", "-c", fmt.Sprintf("gunzip -c %q | cpio -id --quiet", payloadPath))
	cpioCmd.Dir = extractDir
	if out, err := cpioCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("payload extraction failed: %w\n%s", err, out)
	}

	return filepath.Join(extractDir, "usr", "local", "bin", "cierge"), nil
}

// replaceExecutable replaces dst with src. Tries a direct rename first (works
// if the binary is user-owned), and falls back to osascript to trigger macOS's
// native admin credentials dialog if permission is denied (e.g. after a .pkg
// install where /usr/local/bin/cierge is root-owned).
func replaceExecutable(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrPermission) {
		return err
	}

	fmt.Println("Admin privileges required. You will be prompted for your password.")
	script := fmt.Sprintf(
		`do shell script "install -m 755 %q %q" with administrator privileges`,
		src, dst,
	)
	if out, err := exec.Command("osascript", "-e", script).CombinedOutput(); err != nil {
		return fmt.Errorf("privileged install failed: %w\n%s", err, out)
	}

	return nil
}
