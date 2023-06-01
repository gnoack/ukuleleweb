//go:build openbsd

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

func restrictAccess(rwDirs ...string) {
	// Unveil
	for _, path := range rwDirs {
		err := unix.Unveil(path, "rwc")
		if err != nil {
			log.Fatalf(`Unveil(%q, "rwc"): %v`, path, err)
		}
	}
	err := unix.UnveilBlock()
	if err != nil {
		log.Fatalf("UnveilBlock: %v", err)
	}

	// Pledge
	promises := "stdio rpath wpath cpath"
	switch *listenNet {
	case "unix":
		promises += " unix"
	case "tcp", "tcp4", "tcp6":
		promises += " inet"
	default:
		log.Fatalf("unrecognized listen network %q", *listenNet)
	}
	err = unix.Pledge(promises, "")
	if err != nil {
		log.Fatalf(`Pledge(%q, ""): %v`, promises, err)
	}
}
