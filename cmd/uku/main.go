package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "-h", "--help", "help":
		usage()
	case "render":
		runRender(os.Args[2:])
	case "rss":
		runRss(os.Args[2:])
	case "viz":
		runViz(os.Args[2:])
	case "static":
		runStatic(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: uku <subcommand> [flags] [args...]\n\n")
	fmt.Fprintf(os.Stderr, "Subcommands:\n")
	fmt.Fprintf(os.Stderr, "  render  Render markdown files to HTML\n")
	fmt.Fprintf(os.Stderr, "  rss     Generate an RSS feed from wiki pages\n")
	fmt.Fprintf(os.Stderr, "  viz     Visualize the wiki page link graph\n")
	fmt.Fprintf(os.Stderr, "  static  Render wiki pages to a static website\n")
}
