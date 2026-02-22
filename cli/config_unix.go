//go:build !windows

package main

import (
	"fmt"
	"os"
)

// checkConfigPermissions validates that the config file has secure permissions (0600)
func checkConfigPermissions(configPath string) error {
	info, err := os.Stat(configPath)
	if err != nil {
		return err
	}

	if info.Mode().Perm() != 0600 {
		return fmt.Errorf("insecure config file permissions (%v). Run: chmod 0600 %s",
			info.Mode().Perm(), configPath)
	}

	return nil
}
