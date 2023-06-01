//go:build !(linux || openbsd)

package main

func restrictAccess(rwDirs ...string) {
	// Noop
}
