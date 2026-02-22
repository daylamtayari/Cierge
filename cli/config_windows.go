//go:build windows

package main

// Skip check config permissions for Windows
// Could walk the ACLs but will consider that for the future
func checkConfigPermissions(_ string) error {
	return nil
}
